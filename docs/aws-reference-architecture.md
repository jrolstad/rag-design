# AWS Reference Architecture

This document maps the repository's RAG design tenets onto AWS services. The goal is to preserve the architecture's core behavior rather than collapse retrieval and generation into a single managed feature.

## Design Principles

- keep ingestion, retrieval, reranking, and generation loosely coupled
- enforce ACL-aware retrieval before answer generation
- combine lexical and semantic retrieval for technical recall
- prefer canonical documentation over ticket content
- keep context windows small, deduplicated, and citation-backed
- abstain when evidence is weak, inaccessible, or below threshold

## Recommended AWS Service Mapping

| Concern | AWS services | Notes |
| --- | --- | --- |
| Raw document landing zone | Amazon S3 | Store source files, extracted text, and chunk manifests. |
| Ingestion orchestration | AWS Step Functions | Coordinate parse, chunk, embed, index, and retry paths. |
| Parsing and enrichment | AWS Lambda or Amazon ECS/Fargate | Lambda is simpler for moderate volumes. Fargate fits heavy parsing and long-running jobs. |
| Embeddings | Amazon Bedrock | Generate chunk and query embeddings with a managed embeddings model. |
| Hybrid retrieval store | Amazon OpenSearch Serverless | Best default when you want vector and keyword retrieval in one service. |
| Metadata and ACL registry | Amazon DynamoDB or Amazon Aurora PostgreSQL | Store document registry, authority, freshness, and ACL metadata. |
| Query API | Amazon API Gateway plus AWS Lambda | Thin online orchestration layer. |
| Authentication | Amazon Cognito or AWS IAM Identity Center | Pass principal and group claims into retrieval filtering. |
| Answer generation | Amazon Bedrock | Use a text generation model after context assembly. |
| Observability | Amazon CloudWatch and AWS X-Ray | Track ingestion status, retrieval quality, latency, and model usage. |
| Secrets and encryption | AWS Secrets Manager and AWS KMS | Protect credentials and encrypt data at rest. |

## Reference Topology

### Ingestion Path

1. Land raw content in Amazon S3.
2. Trigger an AWS Step Functions state machine on new or changed content.
3. Parse and normalize documents into a canonical shape.
4. Split content into structure-aware chunks.
5. Attach metadata such as `doc_id`, `source_type`, `authority`, `freshness`, `version`, and ACL tags.
6. Generate embeddings with Amazon Bedrock.
7. Index chunks into Amazon OpenSearch Serverless for vector and lexical retrieval.
8. Persist registry and policy metadata in DynamoDB or Aurora PostgreSQL.

### Query Path

1. The client calls an Amazon API Gateway endpoint.
2. A Lambda query service resolves the caller's identity and group claims.
3. The service runs lexical and vector retrieval against OpenSearch Serverless.
4. The service loads ACL and authority metadata from the metadata store.
5. The service filters inaccessible chunks before generation.
6. The service reranks by authority, freshness, and query relevance.
7. The service assembles a small, deduplicated context set.
8. The service calls Amazon Bedrock to generate a citation-backed answer.
9. If the evidence set is empty or below threshold, the service abstains.

## How AWS Supports the Repository Tenets

### Hybrid Retrieval

Use Amazon OpenSearch Serverless as the primary search layer. It supports vector search for semantic retrieval and full-text search for lexical recall, which matches the repository's hybrid retrieval requirement.

### ACL-Aware Retrieval

Do not rely on the LLM layer to hide data. Store ACL metadata with each chunk or document and apply principal-aware filters in the retrieval service before context assembly and generation.

### Authority-Aware Ranking

Model the source authority explicitly. Canonical documentation should carry a stronger authority value than support tickets. Tickets can complement canonical material, but they should not outrank it based only on retrieval score.

### Small, Deduplicated Context Windows

Keep context assembly in the query service. This is where you deduplicate overlapping chunks, enforce token budgets, and select the minimum evidence set needed to answer.

### Citation-Backed Answers and Abstention

Pass source identifiers and chunk references through the full query path so answers can cite the evidence they used. If the filtered evidence set is weak or empty, return an abstention response instead of forcing generation.

## When to Use Bedrock Knowledge Bases

Amazon Bedrock Knowledge Bases can accelerate the first version of the system, especially if you want managed ingestion and retrieval. Use it when speed of delivery matters more than precise control over scoring and policy logic.

For this repository's architecture, a custom retrieval service is still the better fit when you need:

- explicit authority-aware reranking
- strict ACL enforcement before generation
- custom abstention thresholds
- full control over context assembly and citation formatting

## Suggested Build Order

1. Start with S3, Step Functions, Lambda, OpenSearch Serverless, and Bedrock.
2. Implement custom ingestion and metadata enrichment before model tuning.
3. Implement hybrid retrieval plus ACL filtering.
4. Add authority-aware reranking and context assembly.
5. Add answer generation, citations, and abstention behavior.
6. Add evaluation, tracing, and operational dashboards.

## Alternative Storage Option

Use Amazon Aurora PostgreSQL with `pgvector` instead of OpenSearch Serverless when relational joins, transactional updates, or SQL-driven policy logic are more important than a search-native stack. This is a valid alternative, but OpenSearch Serverless is the simpler default for hybrid retrieval.

## Operational Concerns

- use idempotent ingestion steps so reprocessing the same document does not duplicate chunks
- version documents and expire superseded chunks from the retrieval store
- log retrieval candidates, final evidence, and abstention reasons for evaluation
- measure latency separately for retrieval, reranking, and generation
- protect against prompt injection by treating retrieved content as untrusted input
