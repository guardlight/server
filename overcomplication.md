Summary for overcomplexing engineers that needs to keep their job:


At its core, Guardlight operates as a multi-layered, event-driven, containerized microservice architecture designed to facilitate dynamic media content interrogation with asynchronous message-driven processing pipelines.

- UI (User Experience Facilitation Layer)
  The UI, meticulously crafted using React, serves as the primary interaction gateway. It enables bidirectional content ingestion, supporting heterogeneous textual datasets such as EPUB-based literature and manually curated free-text inputs. The system facilitates dynamic theme injection, permitting users to modify lexicon-bound analytical parameters in real-time.

- Server (Centralized Orchestration Nexus)
  The server, a high-throughput Golang-powered execution core, serves as the epicenter of computational governance. It seamlessly integrates with a distributed ACID-compliant relational data store (PostgreSQL) while leveraging containerized execution paradigms via Docker. Furthermore, NATS Jetstream acts as the ephemeral inter-process communication substrate, ensuring atomic message integrity across all analytical subsystems.

- Parsers and Analyzers (Linguistic Deconstruction Engines)
  EPUB Parsing Subsystem: A Python-based lexical extraction engine, optimized for high-fidelity text segmentation, enabling structural decomposition of eBooks into tokenized, analyzable entities.
  Free-Text Parser: A streamlined textual ingestion pipeline that converts unstructured human-generated prose into a format suitable for rigorous computational scrutiny.
  Analysis Computation Unit: The current iteration employs a lexeme-driven probabilistic keyword evaluation mechanism (also known as “word search”).


Distributed Execution and Task Coordination

Guardlight utilizes a heterogeneous task execution paradigm, where a centralized orchestration entity (a.k.a. "Orchestrator") dynamically provisions task-specific containerized execution units in response to computational demand.

- Job Scheduling Algorithm (JSA-9000™): Leveraging an internally managed deterministic job queuing system, the orchestrator distributes workloads via NATS Jetstream’s durable message persistence layer.
- Asynchronous Execution Protocol (AEP): Analytical workloads remain in a transient limbo state within the NATS pipeline until retrieved by an eligible parser/analyzer adhering to the predefined processing contract schema (PPCS).
- Result Post-Processing Pipeline: The system autonomously transmits structured analytical outcomes back into NATS under a dedicated post-analysis data channel, where an additional processing layer consolidates results into a human-consumable interpretative report artifact.
- Front-End Data Synchronization via SSE (Server-Sent Event Manifold™): The UI is perpetually synchronized using unidirectional event streaming, ensuring that the user receives real-time interpretative updates without the necessity for manual refresh cycles.



Final Summary for Non-Engineers:

Guardlight is a React-based UI, a Go server, and some Python parsers that talk to each other using NATS and Docker to process text. You can upload stuff, analyze it, and get a report. That’s it.