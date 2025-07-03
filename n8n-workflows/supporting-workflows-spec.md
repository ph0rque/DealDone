# Supporting Workflows Specification for Enhanced Analyze All

## Overview
This document defines the supporting workflows that complement the main "Enhanced Analyze All" workflow to provide specialized processing capabilities and enhanced error handling.

## 1. Entity Extraction Specialist Workflow

### Purpose
Deep entity analysis for complex documents that require specialized AI processing beyond basic entity extraction.

### Trigger
- **Type**: Webhook
- **Path**: `/entity-extraction-specialist`
- **When**: Called by main workflow when document complexity score > 0.8 or entity extraction confidence < 0.6

### Key Nodes
1. **Complex Document Analyzer**
   - Uses advanced GPT-4 prompts for multi-pass entity extraction
   - Analyzes document structure and relationships
   - Identifies nested financial data and corporate hierarchies

2. **Cross-Reference Validator**
   - Validates entities against external databases
   - Checks company name variations and aliases
   - Verifies financial data consistency

3. **Entity Confidence Booster**
   - Re-processes low-confidence entities with specialized prompts
   - Uses contextual analysis to improve extraction accuracy
   - Provides alternative entity suggestions

### Input Schema
```json
{
  "documentPath": "string",
  "documentType": "string", 
  "basicEntities": "object",
  "complexityScore": "number",
  "dealContext": "object"
}
```

### Output Schema
```json
{
  "enhancedEntities": {
    "companyName": "string",
    "dealValue": "string",
    "revenue": "string",
    "ebitda": "string",
    "keyPersonnel": "array",
    "corporateStructure": "object"
  },
  "confidence": "number",
  "validationResults": "object"
}
```

## 2. Financial Data Validator Workflow

### Purpose
Specialized validation of financial metrics with business rule compliance and ratio analysis.

### Trigger
- **Type**: Webhook  
- **Path**: `/financial-data-validator`
- **When**: Called after field mapping when financial data is detected

### Key Nodes
1. **Financial Data Analyzer**
   - Validates financial ratios and relationships
   - Checks for logical consistency (EBITDA <= Revenue, etc.)
   - Identifies potential data entry errors

2. **Business Rule Validator**
   - Applies M&A-specific business rules
   - Validates against industry benchmarks
   - Flags unusual financial patterns

3. **Currency and Format Standardizer**
   - Standardizes currency formats across documents
   - Handles multiple currency conversions
   - Ensures consistent number formatting

### Input Schema
```json
{
  "extractedFinancials": "object",
  "documentType": "string",
  "industryContext": "string",
  "dealSize": "string"
}
```

### Output Schema
```json
{
  "validatedFinancials": "object",
  "validationResults": {
    "ratioAnalysis": "object",
    "businessRuleCompliance": "boolean",
    "anomalies": "array"
  },
  "formattedData": "object",
  "confidence": "number"
}
```

## 3. Template Quality Assessor Workflow

### Purpose
AI-powered assessment of populated template quality with detailed scoring and recommendations.

### Trigger
- **Type**: Webhook
- **Path**: `/template-quality-assessor`  
- **When**: Called after template population to validate output quality

### Key Nodes
1. **Template Content Analyzer**
   - Reviews populated template for completeness
   - Checks formula integrity and calculations
   - Validates data formatting and presentation

2. **Business Logic Validator**
   - Ensures populated data follows business logic
   - Validates cross-field relationships
   - Checks for missing critical information

3. **Quality Score Calculator**
   - Calculates comprehensive quality metrics
   - Provides improvement recommendations
   - Determines if manual review is needed

### Input Schema
```json
{
  "populatedTemplateIds": "array",
  "originalFieldMappings": "array", 
  "dealContext": "object"
}
```

### Output Schema
```json
{
  "qualityScores": {
    "overall": "number",
    "completeness": "number", 
    "accuracy": "number",
    "formatting": "number"
  },
  "validationResults": "object",
  "recommendations": "array",
  "requiresReview": "boolean"
}
```

## 4. Error Recovery Handler Workflow

### Purpose
Enhanced error handling and recovery with intelligent retry logic and fallback mechanisms.

### Trigger
- **Type**: Webhook
- **Path**: `/error-recovery-handler`
- **When**: Called when any step in main workflow fails

### Key Nodes
1. **Error Classifier**
   - Categorizes error types (API, data, processing, etc.)
   - Determines if error is recoverable
   - Assigns priority level for resolution

2. **Intelligent Retry Manager**
   - Implements exponential backoff for API errors
   - Uses alternative endpoints for service failures
   - Adjusts processing parameters for data errors

3. **Fallback Strategy Executor**
   - Falls back to simpler processing methods
   - Uses cached results when available
   - Provides partial results with quality warnings

4. **Error Notification System**
   - Sends detailed error reports to administrators
   - Creates support tickets for critical failures
   - Provides user-friendly error messages

### Input Schema
```json
{
  "errorType": "string",
  "errorMessage": "string",
  "failedStep": "string",
  "processingContext": "object",
  "retryCount": "number"
}
```

### Output Schema
```json
{
  "recoveryAction": "string",
  "retryRecommended": "boolean",
  "fallbackResult": "object",
  "errorReport": "object",
  "userMessage": "string"
}
```

## Workflow Integration Architecture

### Main Workflow Integration Points
```
Enhanced Analyze All Workflow
    ├── Document Analysis (confidence < 0.6) → Entity Extraction Specialist
    ├── Field Mapping (financial data detected) → Financial Data Validator  
    ├── Template Population (completed) → Template Quality Assessor
    └── Any Step (error occurred) → Error Recovery Handler
```

### Data Flow Between Workflows
1. **Main → Specialist**: Passes context and partial results
2. **Specialist → Main**: Returns enhanced/validated data
3. **Error Handler**: Can intercept and retry any workflow step
4. **Quality Assessor**: Provides final validation before result delivery

## Deployment Strategy

### Phase 1: Core Supporting Workflows
- Deploy Entity Extraction Specialist
- Deploy Error Recovery Handler
- Test integration with main workflow

### Phase 2: Quality Enhancement
- Deploy Financial Data Validator
- Deploy Template Quality Assessor
- Implement comprehensive quality scoring

### Phase 3: Optimization
- Add workflow performance monitoring
- Implement intelligent routing based on document complexity
- Add predictive error prevention

## Monitoring and Metrics

### Key Performance Indicators
- **Entity Extraction Improvement**: Confidence score increase from specialist workflow
- **Financial Validation Accuracy**: Percentage of financial data passing validation
- **Template Quality Scores**: Average quality scores across all populated templates
- **Error Recovery Rate**: Percentage of errors successfully recovered
- **Processing Time Impact**: Additional time added by supporting workflows

### Alerting Thresholds
- Entity extraction specialist called > 30% of the time
- Financial validation failure rate > 10%
- Template quality score < 0.6 for > 20% of templates
- Error recovery failure rate > 5%

This supporting workflow architecture provides specialized processing capabilities while maintaining the efficiency and reliability of the main Enhanced Analyze All workflow. 