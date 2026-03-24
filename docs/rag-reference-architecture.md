# RAG Reference Architecture

This package describes a reference implementation for a retrieval-augmented generation system optimized for internal technical knowledge. The design assumes a mixed corpus of canonical documentation, code-derived references, and support tickets.

## Goals

- favor exact technical recall and semantic recall together
- enforce ACL-aware retrieval before answer generation
- prefer authoritative documentation over ticket content
- abstain when evidence is weak or inaccessible
- keep the retrieval, reranking, and generation stages loosely coupled

## Core Flow

1. Normalize source documents into a canonical `Document` shape.
2. Split documents into structure-aware `Chunk` records.
3. Run hybrid retrieval over lexical and semantic indexes.
4. Filter by ACL and deduplicate candidate evidence.
5. Rerank with authority, freshness, and query relevance.
6. Assemble a small, non-redundant context set.
7. Generate a citation-backed answer or abstain.

## Package Layout

- `internal/rag/models.go`: core types for documents, chunks, evidence, citations, and answers
- `internal/rag/pipeline.go`: reference orchestration pipeline and context assembler
- `internal/rag/pipeline_test.go`: tests for authority preference, ACL filtering, and abstention

## Design Defaults

- canonical docs use `AuthorityCanonical`
- tickets use `AuthoritySecondary`
- tickets can complement docs but must not outrank them solely due to raw retrieval score
- the answer stage returns `Abstained=true` when the final context set is empty or below the minimum evidence threshold

## Extension Points

The reference package is intentionally interface-driven:

- plug in a real query rewriter
- plug in vector and lexical retrievers behind the `Retriever` interface
- plug in a cross-encoder reranker
- plug in an LLM-backed `Generator`

The current implementation is a baseline orchestration layer rather than a production serving stack.
