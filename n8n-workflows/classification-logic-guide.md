# DealDone Document Classification and Routing Logic Guide

This guide explains the intelligent document classification and routing system implemented in n8n workflows for automated M&A document analysis.

## Overview

The classification system uses a hybrid approach combining:
- **Path-based heuristics** for quick initial classification
- **AI-powered analysis** for content-based classification  
- **Composite scoring** for improved accuracy
- **Intelligent routing** to specialized processing paths

## Classification Pipeline Architecture

```
Document Input → Pre-Classification → AI Classification → Composite Scoring → Category Routing → Processing Configuration
```

### 1. Document Pre-Classification

**Node**: `Document Pre-Classifier`

**Purpose**: Analyzes file paths and names to extract context clues before AI processing.

#### Features:
- **File Structure Analysis**: Examines folder hierarchy and file naming patterns
- **Extension Detection**: Identifies file types (PDF, Excel, Word, etc.)
- **Context Extraction**: Captures deal name, folder context, and metadata
- **Initial Hints**: Pattern-based classification hints from file paths

#### Classification Patterns:
- **Financial**: `financial|finance|budget|revenue|profit|loss|income|balance|cash|flow|ebitda|valuation`
- **Legal**: `legal|contract|agreement|terms|conditions|compliance|regulatory|license|permit`
- **Operational**: `operation|process|procedure|workflow|staff|employee|hr|human|resource`
- **Due Diligence**: `due.?diligence|dd|audit|review|assessment|analysis`
- **Technical**: `technical|tech|it|system|software|hardware|infrastructure`
- **Marketing**: `marketing|sales|customer|client|market|brand|promotion`

#### Example Output:
```javascript
{
  filePath: "/deals/AcquisitionCorp/financial/Q3_Financial_Statements.pdf",
  fileName: "Q3_Financial_Statements.pdf",
  fileExtension: "pdf",
  contextClues: {
    folderContext: ["AcquisitionCorp", "financial"],
    dealName: "AcquisitionCorp-TargetInc"
  },
  initialHints: {
    isFinancial: true,
    isLegal: false,
    // ... other categories
  }
}
```

### 2. Document Splitter

**Node**: `Document Splitter`

**Purpose**: Splits batch document processing into individual items for parallel classification.

#### Features:
- **Parallel Processing**: Each document gets its own processing thread
- **Batch Tracking**: Maintains context about the overall batch
- **Index Management**: Preserves document order and relationships

### 3. AI Document Classifier

**Node**: `AI Document Classifier`

**Purpose**: Calls DealDone's AI classification service for content-based analysis.

#### API Endpoint:
- **URL**: `http://localhost:8081/classify-document`
- **Method**: POST
- **Authentication**: API Key via X-API-Key header

#### Request Payload:
```json
{
  "filePath": "/path/to/document.pdf",
  "fileName": "document.pdf", 
  "dealName": "DealName",
  "jobId": "job-12345",
  "contextClues": { /* file context */ },
  "initialHints": { /* path-based hints */ },
  "classificationDepth": "comprehensive"
}
```

#### Expected Response:
```json
{
  "primaryCategory": "financial",
  "confidence": 0.87,
  "reasoning": "Document contains financial statements and cash flow analysis",
  "categories": {
    "financial": 0.87,
    "legal": 0.12,
    "operational": 0.05
  },
  "keywords": ["revenue", "profit", "balance sheet"],
  "summary": "Q3 financial statements with revenue analysis"
}
```

### 4. Classification Processor

**Node**: `Classification Processor`

**Purpose**: Combines AI results with path-based hints using composite scoring.

#### Composite Scoring Algorithm:
```javascript
compositeScore = (aiScore * 0.7) + (pathHint * 0.3)
```

- **AI Weight**: 70% - Content-based analysis
- **Path Weight**: 30% - File structure heuristics

#### Classification Logic:
1. **Extract AI Results**: Primary category, confidence, reasoning
2. **Calculate Composite Scores**: For each category using weighted formula
3. **Determine Final Classification**: Highest composite score wins
4. **Set Confidence Levels**:
   - High: ≥ 0.8
   - Medium: 0.5 - 0.79
   - Low: < 0.5
   - Manual Review Required: < 0.3

#### Example Processing:
```javascript
// AI says 75% financial, path analysis confirms financial folder
compositeScoring = {
  financial: (0.75 * 0.7) + (1.0 * 0.3) = 0.825,  // High confidence
  legal: (0.15 * 0.7) + (0.0 * 0.3) = 0.105,      // Low
  operational: (0.10 * 0.7) + (0.0 * 0.3) = 0.070 // Very low
}

finalClassification = {
  primaryCategory: "financial",
  primaryScore: 0.825,
  confidence: 0.75,
  isHighConfidence: true
}
```

### 5. Category Router

**Node**: `Category Router`

**Purpose**: Routes documents to specialized processing paths based on classification.

#### Routing Conditions:
- **Financial Route**: `primaryCategory === "financial"`
- **Legal Route**: `primaryCategory === "legal"`  
- **Operational Route**: `primaryCategory === "operational"`
- **Due Diligence Route**: `primaryCategory === "dueDiligence"`
- **Technical Route**: `primaryCategory === "technical"`
- **Marketing Route**: `primaryCategory === "marketing"`
- **General Route**: All other cases (fallback)

### 6. Specialized Processing Routers

Each category has its own specialized processing configuration:

#### Financial Router
```javascript
{
  processingType: 'financial-enhanced',
  extractors: [
    'revenue-analyzer',
    'balance-sheet-parser', 
    'cash-flow-analyzer',
    'valuation-calculator',
    'financial-ratio-computer'
  ],
  templates: ['financial-model', 'valuation-template'],
  confidenceThreshold: 0.85,
  estimatedTime: 120000 // 2 minutes
}
```

#### Legal Router
```javascript
{
  processingType: 'legal-enhanced',
  extractors: [
    'contract-analyzer',
    'legal-entity-extractor',
    'compliance-checker',
    'terms-condition-parser'
  ],
  templates: ['legal-summary', 'compliance-checklist'],
  confidenceThreshold: 0.90,
  estimatedTime: 180000 // 3 minutes
}
```

#### General Router (Fallback)
```javascript
{
  processingType: 'general-standard',
  extractors: [
    'general-content-extractor',
    'keyword-analyzer',
    'document-summarizer'
  ],
  templates: ['general-summary', 'document-overview'],
  confidenceThreshold: 0.60,
  estimatedTime: 60000 // 1 minute
}
```

## Advanced Features

### Confidence-Based Processing

Documents are processed differently based on classification confidence:

- **High Confidence (≥0.8)**: Fast-track processing with standard validation
- **Medium Confidence (0.5-0.79)**: Enhanced validation and quality checks
- **Low Confidence (<0.5)**: Thorough validation and potential manual review
- **Very Low (<0.3)**: Automatic manual review flag

### Manual Review System

Low confidence classifications trigger a manual review process:

```javascript
manualReviewRequest = {
  documentPath: "/path/to/document.pdf",
  classificationAttempt: { /* AI and composite results */ },
  reviewReason: "low-confidence",
  suggestedCategories: [/* top 3 possibilities */],
  reviewPriority: "normal|high",
  estimatedReviewTime: 300000, // 5 minutes
  fallbackProcessing: "general-with-flags"
}
```

### Processing Path Determination

Each document gets a specific processing path:

1. **Classification Category**: Financial, Legal, Operational, etc.
2. **Confidence Level**: High, Medium, Low
3. **Processing Type**: Enhanced, Standard, Manual Review
4. **Resource Allocation**: Premium, Standard, Basic
5. **Quality Validation**: Minimal, Standard, Thorough

## Integration with DealDone

### Required DealDone API Endpoints

The classification system expects these endpoints in DealDone:

1. **`/classify-document`**: AI-powered document classification
2. **`/analyze-document`**: Content analysis and data extraction
3. **`/extract-financial-data`**: Specialized financial analysis
4. **`/extract-legal-entities`**: Legal document processing
5. **`/discover-templates`**: Template matching and discovery

### Authentication Requirements

All API calls use:
- **Header**: `X-API-Key: {generated-api-key}`
- **Content-Type**: `application/json`
- **Request ID**: `X-Request-ID` for tracking

## Performance Optimization

### Parallel Processing
- Documents are processed simultaneously
- Each classification runs independently
- Results are aggregated after routing

### Caching Strategy
- Classification results can be cached
- Path-based hints are reusable
- Template configurations are shared

### Resource Management
- High-priority documents get more resources
- Processing time estimates guide scheduling
- Timeout values are category-specific

## Error Handling

### Classification Failures
1. **AI Service Unavailable**: Fall back to path-based classification
2. **Low Confidence Results**: Route to manual review
3. **Network Timeouts**: Retry with exponential backoff
4. **Invalid Responses**: Log error and use general processing

### Recovery Mechanisms
- Automatic retry for transient failures
- Graceful degradation to simpler classification
- Manual review queue for problematic documents
- Audit logging for all classification decisions

## Monitoring and Analytics

### Classification Metrics
- **Accuracy**: Confidence score distributions
- **Performance**: Processing times per category
- **Volume**: Documents processed by type
- **Quality**: Manual review rates

### Key Performance Indicators
- Average classification confidence: >0.75
- Manual review rate: <10%
- Processing time per document: <2 minutes
- Classification accuracy (when validated): >95%

## Configuration and Tuning

### Adjustable Parameters
- **Composite Score Weights**: AI vs Path analysis (default 70/30)
- **Confidence Thresholds**: High/Medium/Low boundaries
- **Processing Timeouts**: Per category time limits
- **Retry Logic**: Failure handling parameters

### A/B Testing Support
- Multiple classification models can be tested
- Performance comparison between approaches
- Gradual rollout of classification improvements

## Best Practices

### File Organization
- Use consistent folder naming conventions
- Include category keywords in file paths
- Maintain clear deal folder structures

### Document Naming
- Include descriptive keywords in filenames
- Use standardized naming patterns
- Avoid special characters and spaces

### Quality Assurance
- Regularly review manual classification requests
- Monitor classification accuracy metrics
- Update classification patterns based on feedback

## Troubleshooting

### Common Issues
1. **Misclassification**: Check file path patterns and AI training data
2. **Low Confidence**: Review document quality and content clarity
3. **Processing Delays**: Monitor AI service performance and timeouts
4. **Route Failures**: Verify category router conditions and connections

### Debug Information
- Classification reasoning from AI service
- Composite score calculations
- Path-based hint analysis
- Processing configuration details

This classification and routing system provides the foundation for intelligent, automated document processing in M&A workflows, ensuring documents are routed to the most appropriate processing paths for optimal results. 