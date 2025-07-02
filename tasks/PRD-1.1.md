# Product Requirements Document: n8n Workflow Integration for Document Analysis

## Introduction/Overview

The n8n Workflow Integration feature extends DealDone's existing document analysis capabilities by implementing a robust, queue-based workflow system that handles document processing through n8n automation platform. This feature moves the core document classification and analysis mechanism from direct processing to a sophisticated workflow-driven approach that provides better scalability, error handling, and state management.

The system will use push-based webhooks where DealDone triggers n8n workflows when files change or users request analysis, with n8n handling the heavy lifting of document processing, template population, and intelligent data merging before returning results back to DealDone.

## Goals

1. **Implement robust workflow-based document processing** using n8n to replace direct AI processing calls
2. **Establish reliable push-based architecture** where DealDone triggers n8n workflows via webhooks
3. **Create intelligent queueing system** that processes documents sequentially while handling batch operations
4. **Build resilient error handling** with automatic retries, partial completion support, and workflow resumption
5. **Implement intelligent data merging** with composite confidence scoring when multiple documents populate the same template fields
6. **Maintain state consistency** between DealDone file system and n8n workflow tracking
7. **Enable seamless background processing** for long-running AI analysis jobs

## User Stories

1. **As a deal analyst**, I want to drop documents into a deal folder and have them automatically trigger n8n workflows for processing, so that analysis happens reliably in the background without blocking my work.

2. **As a financial analyst**, I want multiple financial documents to be intelligently merged into my templates with composite confidence scores, so that I get the most accurate data from all available sources.

3. **As a deal manager**, I want processing failures to be handled gracefully with automatic retries, so that temporary issues don't prevent deal analysis from completing.

4. **As a system administrator**, I want n8n to maintain its own tracking of deal folder structures and processing state, so that workflows can resume properly after failures.

5. **As a deal team member**, I want the system to queue document processing requests intelligently, so that simultaneous uploads don't cause conflicts or data corruption.

6. **As a senior analyst**, I want user corrections to flow back through the n8n workflow for RAG-based learning, so that the system continuously improves its accuracy.

## Functional Requirements

### 1. Webhook Integration
- 1.1 DealDone must expose webhook endpoints that n8n can call for receiving processing results
- 1.2 DealDone must send webhook payloads to n8n when files are added to deal folders
- 1.3 DealDone must send webhook payloads to n8n when users click "Analyze" or "Analyze All" buttons
- 1.4 Webhook payloads must include: dealName, filePaths array, triggerType, jobID, and timestamp
- 1.5 All webhook communications must use JSON format with proper error handling

### 2. Queue Management
- 2.1 n8n workflows must implement a processing queue that handles one document analysis at a time
- 2.2 The queue must support batch processing where batch size can be 1 for single files
- 2.3 Queue items must include job metadata: ID, deal name, file paths, status, retry count, timestamps
- 2.4 The system must prevent race conditions when multiple files are dropped simultaneously
- 2.5 Queue status must be queryable by DealDone for progress tracking

### 3. Document Classification and Routing
- 3.1 n8n workflows must call DealDone's existing document classification APIs
- 3.2 Classified documents must be tracked in n8n's mirror of the deal folder structure
- 3.3 If a deal folder doesn't exist, n8n must trigger DealDone to recreate it
- 3.4 Only individual file drops are supported (not nested folder drops)
- 3.5 Document classification results must include confidence scores and keywords

### 4. Intelligent Data Merging
- 4.1 When multiple documents map to the same template fields, the system must implement composite confidence scoring
- 4.2 Higher confidence data must override lower confidence data in template population
- 4.3 Equal confidence numeric data must be averaged with appropriate notation
- 4.4 Conflict resolution must maintain history of previous values and their confidence levels
- 4.5 Template field conflicts must be tracked and queryable for audit purposes

### 5. Template Discovery and Population
- 5.1 n8n workflows must discover relevant templates based on document type classification
- 5.2 Field mapping must use DealDone's existing intelligent matching algorithms (exact, synonym, fuzzy)
- 5.3 Template population must preserve Excel formulas and formatting
- 5.4 Populated templates must be saved to the deal's analysis folder
- 5.5 Population results must include success/failure status for each template and field

### 6. Error Handling and Recovery
- 6.1 Failed workflows must be retryable with exponential backoff
- 6.2 Partial processing completion must be supported (some documents succeed, others fail)
- 6.3 AI service failures must be handled gracefully with appropriate fallback mechanisms
- 6.4 Failed workflows must be resumable from the last successful step
- 6.5 Error details must be logged and queryable by DealDone

### 7. State Management and Tracking
- 7.1 n8n must maintain a mirror of deal folder structures for workflow state tracking
- 7.2 Processing history must be maintained for each document and template
- 7.3 Workflow state must persist across n8n restarts and failures
- 7.4 Deal structure changes must be synchronized between DealDone and n8n
- 7.5 Processing results must be stored in appropriate analysis folder files

### 8. User Correction Integration
- 8.1 User corrections to template data must be detected and sent to n8n workflows
- 8.2 Corrections must be processed through RAG-based learning mechanisms
- 8.3 Learning feedback must be integrated into the main document processing workflow
- 8.4 Correction history must be maintained for audit and improvement tracking
- 8.5 Updated confidence models must be applied to future document processing

## Non-Goals (Out of Scope)

1. This feature will NOT support nested folder drops or complex folder hierarchies
2. This feature will NOT implement concurrent document processing (sequential only)
3. This feature will NOT include rate limiting between DealDone and n8n
4. This feature will NOT support real-time collaborative editing of templates
5. This feature will NOT implement custom n8n node development (uses existing nodes)
6. This feature will NOT support multiple n8n instance load balancing
7. This feature will NOT include advanced workflow analytics beyond basic tracking

## Design Considerations

### 1. Workflow Architecture
- **Primary Workflow**: `dealdone-document-processor` handles main processing pipeline
- **Supporting Workflows**: Separate workflows for error handling, user corrections, and cleanup
- **Node Types**: HTTP Request nodes for DealDone API calls, Function nodes for logic, Webhook nodes for triggers
- **Data Flow**: Push-based with DealDone initiating all workflow triggers

### 2. State Management
- **Global Workflow Data**: Use n8n's workflow static data for queue and state persistence
- **Deal Structure Mirror**: Maintain JSON representation of deal folders in n8n
- **Conflict Tracking**: Store field-level conflicts with timestamps and confidence history
- **Processing History**: Track all document processing attempts with results and errors

### 3. Error Handling Strategy
- **Retry Logic**: Exponential backoff with maximum retry limits
- **Partial Success**: Mark individual documents as complete/failed within batches
- **Graceful Degradation**: Continue processing remaining documents when one fails
- **Recovery Points**: Allow workflow resumption from specific steps

## Technical Considerations

### 1. API Integration
- **New DealDone Endpoints**: Add webhook receivers, status queries, and result handlers
- **Authentication**: Implement secure API key exchange between DealDone and n8n
- **Payload Structure**: Standardize JSON schemas for all webhook communications
- **Timeout Handling**: Configure appropriate timeouts for long-running AI operations

### 2. Performance Optimization
- **Queue Efficiency**: Implement FIFO queue with priority support for user-triggered requests
- **Memory Management**: Clean up completed job data to prevent memory leaks
- **Batch Optimization**: Group related documents for more efficient processing
- **Caching Strategy**: Leverage existing DealDone caching for repeated operations

### 3. Data Consistency
- **Atomic Operations**: Ensure template updates are atomic to prevent corruption
- **State Synchronization**: Regular sync checks between DealDone and n8n state
- **Conflict Resolution**: Clear precedence rules for conflicting data sources
- **Audit Trail**: Complete logging of all data changes and their sources

## Success Metrics

1. **Reliability**: 99%+ workflow completion rate with automatic error recovery
2. **Processing Time**: Average document processing time under 2 minutes per document
3. **Data Accuracy**: Maintain current 95%+ accuracy in document classification
4. **Queue Efficiency**: Zero queue backlogs under normal operating conditions
5. **Error Recovery**: 90%+ of failed workflows successfully resume without manual intervention
6. **User Satisfaction**: Zero reported data corruption or template population errors
7. **System Uptime**: n8n workflow system maintains 99.9% availability

## Implementation Phases

### Phase 1: Basic Integration (Week 1-2)
- Implement webhook endpoints in DealDone
- Create basic n8n workflow for single document processing
- Implement queue management system
- Add file change trigger mechanisms

### Phase 2: Advanced Features (Week 3-4)
- Implement intelligent data merging logic
- Add comprehensive error handling and retry mechanisms
- Create user correction feedback workflows
- Implement state tracking and folder structure mirroring

### Phase 3: Optimization and Polish (Week 5+)
- Performance tuning and optimization
- Enhanced error recovery mechanisms
- Advanced conflict resolution algorithms
- Comprehensive testing and monitoring

## Open Questions

1. **n8n Hosting**: Should we use n8n Cloud or self-hosted instance for better control?
2. **Webhook Security**: What authentication mechanism should we use for webhook communications?
3. **Queue Persistence**: Should the processing queue persist across n8n restarts?
4. **Monitoring**: What level of workflow monitoring and alerting do we need?
5. **Backup Strategy**: How should we handle n8n workflow backup and disaster recovery?
6. **Scaling**: At what point would we need to consider multiple n8n instances?
7. **Testing**: How do we effectively test complex workflow scenarios in development?

## Dependencies

1. **n8n Platform**: Requires n8n instance (cloud or self-hosted) with webhook capabilities
2. **DealDone APIs**: Extends existing API surface with new webhook and status endpoints  
3. **AI Services**: Continues to use existing OpenAI/Claude integration through DealDone
4. **File System**: Maintains current file system-based architecture
5. **Template System**: Builds on existing template parsing and population capabilities 