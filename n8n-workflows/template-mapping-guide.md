# Template Discovery and Field Mapping Guide

This guide explains the intelligent template discovery and field mapping system for M&A document analysis.

## Overview

The template discovery and field mapping system:
- Discovers relevant templates based on document classification
- Extracts structured data from documents
- Maps document fields to template fields with confidence scoring
- Prepares documents for template population

## Pipeline Architecture

```
Classified Document → Template Discovery → Field Extraction → Field Mapping → Quality Validation
```

## Template Discovery Process

### 1. Template Categories by Document Type

- **Financial**: financial-model, valuation-template, due-diligence-financial, cash-flow-analysis
- **Legal**: legal-contract-summary, compliance-checklist, legal-risk-matrix, regulatory-overview
- **Operational**: operational-overview, process-documentation, resource-planning, operational-metrics
- **Due Diligence**: due-diligence-summary, risk-assessment-matrix, compliance-status-report
- **Technical**: technical-overview, system-architecture-diagram, technology-assessment
- **Marketing**: market-analysis-summary, competitive-landscape-overview, customer-segment-analysis
- **General**: general-document-summary, content-analysis-report, document-overview

### 2. Discovery Parameters

```javascript
const discoveryParams = {
  primaryCategory: classification.primaryCategory,
  confidence: classification.confidence,
  documentType: doc.fileExtension,
  dealName: payload.dealName,
  templateCategories: getTemplateCategories(classification.primaryCategory),
  extractedKeywords: doc.aiClassification?.extractedKeywords || []
};
```

## Field Mapping Process

### 1. Field Extraction

The system extracts structured data from documents using:
- AI-powered content analysis
- Pattern matching algorithms
- Named entity recognition
- Table and form data extraction

### 2. Template Field Mapping

Maps extracted fields to template fields using:
- **Direct Matching**: Exact field name matches
- **Semantic Matching**: AI-powered similarity analysis
- **Fuzzy Matching**: Approximate string matching
- **Pattern Matching**: Regular expression patterns

### 3. Quality Assessment

```javascript
const qualityAnalysis = {
  totalMappings: fieldMappings.length,
  successfulMappings: fieldMappings.filter(m => m.confidence >= 0.7).length,
  averageConfidence: /* calculated average */,
  overallScore: /* success rate */
};

const readyForPopulation = qualityAnalysis.overallScore >= 0.6;
```

## API Integration

### Required DealDone Endpoints

1. **`/discover-templates`** - Template discovery and matching
2. **`/extract-document-fields`** - Document field extraction
3. **`/map-template-fields`** - Intelligent field mapping

### Request/Response Examples

#### Template Discovery Request
```json
{
  "discoveryParams": {
    "primaryCategory": "financial",
    "confidence": 0.85,
    "documentType": "pdf",
    "templateCategories": ["financial-model", "valuation-template"],
    "extractedKeywords": ["revenue", "profit", "cash flow"]
  },
  "dealName": "AcquisitionCorp-TargetInc",
  "jobId": "job-12345"
}
```

#### Template Discovery Response
```json
{
  "templateMatches": [
    {
      "templateId": "template-001",
      "name": "Financial Model Template",
      "matchScore": 0.92,
      "fields": ["revenue", "expenses", "profit"],
      "requiredFields": ["revenue", "expenses"]
    }
  ]
}
```

#### Field Extraction Response
```json
{
  "extractedFields": {
    "revenue": {
      "value": "50000000",
      "confidence": 0.95,
      "source": "financial_statement_page_2",
      "dataType": "currency"
    }
  }
}
```

#### Field Mapping Response
```json
{
  "mappings": [
    {
      "templateField": "revenue",
      "documentField": "total_revenue", 
      "value": "50000000",
      "confidence": 0.95,
      "mappingType": "direct_match"
    }
  ]
}
```

## Quality Metrics

### Template Discovery
- **Discovery Success Rate**: Documents with suitable templates found
- **Template Match Accuracy**: Quality of template recommendations
- **Processing Performance**: Average discovery time

### Field Mapping
- **Mapping Success Rate**: Fields successfully mapped
- **Mapping Confidence**: Average confidence scores
- **Data Quality**: Valid, complete field values

## Error Handling

### No Templates Found
- Route to generic extraction fallback
- Flag for manual template selection
- Create template creation recommendation

### Poor Field Extraction
- Retry with simplified parameters
- Route to manual data entry
- Apply OCR preprocessing if needed

### Low Mapping Quality
- Require manual validation
- Flag incomplete required fields
- Apply data transformation

## Best Practices

### Template Design
- Use clear, unambiguous field names
- Maintain consistent naming conventions
- Provide comprehensive template metadata

### Document Quality
- Ensure consistent formatting
- Use well-structured layouts
- Provide high-resolution, searchable files

### Performance Optimization
- Cache template discovery results
- Use parallel processing where possible
- Implement appropriate timeout values

## Troubleshooting

### Common Issues
1. **No Templates Found**: Expand template library or use generic processing
2. **Poor Extraction**: Improve document quality or use manual entry
3. **Low Confidence**: Validate mappings manually or improve algorithms
4. **Performance Issues**: Optimize processing or segment large documents

This system provides intelligent automation for template discovery and field mapping, significantly reducing manual effort in M&A document processing while maintaining high accuracy and quality standards. 