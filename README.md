# rag-design

Reference architecture and starter implementation for a retrieval-augmented generation system optimized for internal technical knowledge.

## Contents

- `docs/rag-reference-architecture.md`: architecture summary and design defaults
- `docs/diagrams.md`: one-page Mermaid diagrams
- `internal/rag/models.go`: core contracts for documents, chunks, evidence, citations, and answers
- `internal/rag/pipeline.go`: reference orchestration pipeline with ACL filtering, authority-aware selection, and abstention
- `internal/rag/pipeline_test.go`: tests for core RAG behaviors
- `go.mod`: minimal Go module for the reference package

## Key Defaults

- hybrid retrieval: keyword + vector
- ACL filtering before answer generation
- canonical docs outrank tickets
- small, deduplicated context windows
- citations required for grounded answers
- abstain when evidence is insufficient
