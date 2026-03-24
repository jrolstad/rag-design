package rag

import (
	"context"
	"testing"
)

type stubRetriever struct {
	results []RankedEvidence
}

func (s stubRetriever) Retrieve(_ context.Context, _ QueryRequest) ([]RankedEvidence, error) {
	return s.results, nil
}

func TestContextAssemblerPrefersCanonicalSourcesOverTickets(t *testing.T) {
	assembler := ContextAssembler{MaxChunks: 2}

	ticket := RankedEvidence{
		Chunk: Chunk{
			ID: "ticket-1",
			Document: Document{
				SourceID:       "ticket-doc",
				SourceType:     SourceTypeTicket,
				AuthorityLevel: AuthoritySecondary,
			},
		},
		Score: 0.99,
	}

	doc := RankedEvidence{
		Chunk: Chunk{
			ID: "doc-1",
			Document: Document{
				SourceID:       "doc-source",
				SourceType:     SourceTypeDocumentation,
				AuthorityLevel: AuthorityCanonical,
			},
		},
		Score: 0.80,
	}

	result := assembler.Assemble([]RankedEvidence{ticket, doc})
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].Chunk.Document.SourceType != SourceTypeDocumentation {
		t.Fatalf("expected canonical documentation first, got %s", result[0].Chunk.Document.SourceType)
	}
}

func TestPipelineAbstainsWhenAclRemovesAllEvidence(t *testing.T) {
	pipeline := Pipeline{
		Retriever: stubRetriever{
			results: []RankedEvidence{
				{
					Chunk: Chunk{
						ID: "restricted-doc",
						Document: Document{
							SourceID:       "doc-source",
							Title:          "Restricted Runbook",
							ACL:            []string{"sre"},
							AuthorityLevel: AuthorityCanonical,
						},
					},
					Score: 0.9,
				},
			},
		},
		MinContextChunks: 1,
	}

	response, err := pipeline.Answer(context.Background(), QueryRequest{
		Query:      "How does failover work?",
		UserGroups: []string{"engineering"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !response.Abstained {
		t.Fatalf("expected response to abstain")
	}
	if response.Diagnostics.AbstainReason != "insufficient_grounded_context" {
		t.Fatalf("unexpected abstain reason: %s", response.Diagnostics.AbstainReason)
	}
}

func TestPipelineReturnsCitationsForGroundedEvidence(t *testing.T) {
	pipeline := Pipeline{
		Retriever: stubRetriever{
			results: []RankedEvidence{
				{
					Chunk: Chunk{
						ID: "doc-1",
						Document: Document{
							SourceID:       "api-errors",
							Title:          "API Error Catalog",
							SourceType:     SourceTypeDocumentation,
							AuthorityLevel: AuthorityCanonical,
							Version:        "v2",
						},
						ContentHash: "hash-1",
					},
					Score: 0.95,
				},
			},
		},
		MinContextChunks: 1,
	}

	response, err := pipeline.Answer(context.Background(), QueryRequest{
		Query:      "What does ERR-42 mean?",
		UserGroups: []string{"engineering"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Abstained {
		t.Fatalf("expected grounded answer")
	}
	if len(response.Citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(response.Citations))
	}
	if response.Citations[0].SourceID != "api-errors" {
		t.Fatalf("unexpected citation source: %s", response.Citations[0].SourceID)
	}
}
