package rag

import "time"

type SourceType string

const (
	SourceTypeDocumentation SourceType = "documentation"
	SourceTypeCodeReference SourceType = "code_reference"
	SourceTypeTicket        SourceType = "ticket"
)

type AuthorityLevel int

const (
	AuthorityUnknown AuthorityLevel = iota
	AuthoritySecondary
	AuthorityCanonical
)

type Document struct {
	SourceID       string
	SourceType     SourceType
	Title          string
	Content        string
	SectionPath    []string
	ComponentTags  []string
	Version        string
	UpdatedAt      time.Time
	ACL            []string
	AuthorityLevel AuthorityLevel
	CanonicalID    string
}

type Chunk struct {
	ID           string
	Document     Document
	Text         string
	ChunkType    string
	TokenCount   int
	StartOffset  int
	EndOffset    int
	LexicalTerms []string
	EmbeddingRef string
	ContentHash  string
}

type QueryRequest struct {
	Query      string
	UserID     string
	UserGroups []string
	Limit      int
}

type RankedEvidence struct {
	Chunk         Chunk
	Score         float64
	RetrievalMode string
	Reason        string
}

type Citation struct {
	SourceID    string
	Title       string
	ChunkID     string
	SourceType  SourceType
	SectionPath []string
	Version     string
}

type Diagnostics struct {
	RewrittenQuery    string
	RetrievedCount    int
	ContextChunkCount int
	AbstainReason     string
}

type AnswerResponse struct {
	Answer      string
	Citations   []Citation
	Abstained   bool
	FollowUp    string
	Diagnostics Diagnostics
}
