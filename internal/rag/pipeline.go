package rag

import (
	"context"
	"fmt"
	"sort"
)

type QueryRewriter interface {
	Rewrite(ctx context.Context, request QueryRequest) (string, error)
}

type Retriever interface {
	Retrieve(ctx context.Context, request QueryRequest) ([]RankedEvidence, error)
}

type Reranker interface {
	Rerank(ctx context.Context, query string, candidates []RankedEvidence) ([]RankedEvidence, error)
}

type Generator interface {
	Generate(ctx context.Context, query string, context []RankedEvidence) (AnswerResponse, error)
}

type Pipeline struct {
	Rewriter            QueryRewriter
	Retriever           Retriever
	Reranker            Reranker
	Generator           Generator
	ContextAssembler    ContextAssembler
	MinContextChunks    int
	DefaultRetrieveSize int
}

type ContextAssembler struct {
	MaxChunks int
}

func (p Pipeline) Answer(ctx context.Context, request QueryRequest) (AnswerResponse, error) {
	query := request.Query
	if request.Limit == 0 {
		request.Limit = p.DefaultRetrieveSize
	}
	if request.Limit == 0 {
		request.Limit = 12
	}

	if p.Rewriter != nil {
		rewritten, err := p.Rewriter.Rewrite(ctx, request)
		if err != nil {
			return AnswerResponse{}, err
		}
		if rewritten != "" {
			query = rewritten
			request.Query = rewritten
		}
	}

	candidates, err := p.Retriever.Retrieve(ctx, request)
	if err != nil {
		return AnswerResponse{}, err
	}

	filtered := filterAuthorizedEvidence(candidates, request.UserGroups)
	filtered = dedupeEvidence(filtered)

	if p.Reranker != nil {
		filtered, err = p.Reranker.Rerank(ctx, query, filtered)
		if err != nil {
			return AnswerResponse{}, err
		}
	}

	assembler := p.ContextAssembler
	if assembler.MaxChunks == 0 {
		assembler.MaxChunks = 6
	}

	contextSet := assembler.Assemble(filtered)
	response := AnswerResponse{
		Diagnostics: Diagnostics{
			RewrittenQuery:    query,
			RetrievedCount:    len(filtered),
			ContextChunkCount: len(contextSet),
		},
	}

	if len(contextSet) < maxInt(1, p.MinContextChunks) {
		response.Abstained = true
		response.Answer = "I do not have enough grounded evidence to answer that reliably."
		response.FollowUp = "Refine the query with a product, module, version, or exact error identifier."
		response.Diagnostics.AbstainReason = "insufficient_grounded_context"
		return response, nil
	}

	if p.Generator == nil {
		response.Answer = fmt.Sprintf("Found %d grounded source chunks for query %q.", len(contextSet), query)
		response.Citations = citationsFromEvidence(contextSet)
		return response, nil
	}

	generated, err := p.Generator.Generate(ctx, query, contextSet)
	if err != nil {
		return AnswerResponse{}, err
	}

	generated.Diagnostics.RewrittenQuery = query
	generated.Diagnostics.RetrievedCount = len(filtered)
	generated.Diagnostics.ContextChunkCount = len(contextSet)
	if len(generated.Citations) == 0 {
		generated.Citations = citationsFromEvidence(contextSet)
	}
	return generated, nil
}

func (a ContextAssembler) Assemble(candidates []RankedEvidence) []RankedEvidence {
	if a.MaxChunks == 0 {
		a.MaxChunks = 6
	}

	sorted := append([]RankedEvidence(nil), candidates...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := sorted[i]
		right := sorted[j]

		if left.Chunk.Document.AuthorityLevel != right.Chunk.Document.AuthorityLevel {
			return left.Chunk.Document.AuthorityLevel > right.Chunk.Document.AuthorityLevel
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		return left.Chunk.ID < right.Chunk.ID
	})

	result := make([]RankedEvidence, 0, minInt(a.MaxChunks, len(sorted)))
	seenHashes := make(map[string]struct{}, len(sorted))
	seenDocs := make(map[string]int, len(sorted))

	for _, candidate := range sorted {
		if len(result) == a.MaxChunks {
			break
		}
		if candidate.Chunk.ContentHash != "" {
			if _, exists := seenHashes[candidate.Chunk.ContentHash]; exists {
				continue
			}
		}

		docID := candidate.Chunk.Document.SourceID
		if seenDocs[docID] >= 2 {
			continue
		}

		result = append(result, candidate)
		seenDocs[docID]++
		if candidate.Chunk.ContentHash != "" {
			seenHashes[candidate.Chunk.ContentHash] = struct{}{}
		}
	}

	return result
}

func filterAuthorizedEvidence(candidates []RankedEvidence, userGroups []string) []RankedEvidence {
	if len(candidates) == 0 {
		return nil
	}

	allowedGroups := make(map[string]struct{}, len(userGroups))
	for _, group := range userGroups {
		allowedGroups[group] = struct{}{}
	}

	filtered := make([]RankedEvidence, 0, len(candidates))
	for _, candidate := range candidates {
		if aclPermits(candidate.Chunk.Document.ACL, allowedGroups) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func aclPermits(documentACL []string, allowedGroups map[string]struct{}) bool {
	if len(documentACL) == 0 {
		return true
	}
	for _, acl := range documentACL {
		if _, ok := allowedGroups[acl]; ok {
			return true
		}
	}
	return false
}

func dedupeEvidence(candidates []RankedEvidence) []RankedEvidence {
	seen := make(map[string]struct{}, len(candidates))
	result := make([]RankedEvidence, 0, len(candidates))
	for _, candidate := range candidates {
		if _, exists := seen[candidate.Chunk.ID]; exists {
			continue
		}
		seen[candidate.Chunk.ID] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

func citationsFromEvidence(evidence []RankedEvidence) []Citation {
	citations := make([]Citation, 0, len(evidence))
	for _, item := range evidence {
		citations = append(citations, Citation{
			SourceID:    item.Chunk.Document.SourceID,
			Title:       item.Chunk.Document.Title,
			ChunkID:     item.Chunk.ID,
			SourceType:  item.Chunk.Document.SourceType,
			SectionPath: item.Chunk.Document.SectionPath,
			Version:     item.Chunk.Document.Version,
		})
	}
	return citations
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
