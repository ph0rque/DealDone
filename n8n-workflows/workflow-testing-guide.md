# DealDone n8n Workflow Testing Guide

## Overview

This guide provides comprehensive testing procedures for all DealDone n8n workflows. Use this guide to validate workflow functionality, debug configuration issues, and ensure seamless integration with the DealDone application.

## Prerequisites

### n8n Environment Setup
1. **n8n Installation**: Version 1.0+ with Node.js support
2. **Database**: PostgreSQL or MySQL for production persistence
3. **Network Access**: Connectivity to DealDone application (localhost:8081)
4. **Credentials**: DealDone API keys and webhook authentication

### DealDone Application Requirements
1. **Webhook Service**: Running on port 8081
2. **Authentication Manager**: API key generation active
3. **Job Tracker**: Persistent job storage enabled
4. **Document Processing**: AI services configured

## Workflow Import Procedure

### 1. Import Workflow Files
```bash
# Copy workflow files to n8n import directory
cp dealdone-document-processor.json /n8n/workflows/
cp dealdone-error-handler.json /n8n/workflows/
cp dealdone-user-corrections.json /n8n/workflows/
cp dealdone-cleanup.json /n8n/workflows/
```

### 2. Configure Credentials
Create the following credential sets in n8n:

#### API Key Credential (dealdone-api-key)
```json
{
  "name": "DealDone API Key",
  "type": "httpHeaderAuth",
  "data": {
    "name": "X-API-Key",
    "value": "your-dealdone-api-key-here"
  }
}
```

#### HMAC Signature Credential (dealdone-hmac)
```json
{
  "name": "DealDone HMAC",
  "type": "httpHeaderAuth",
  "data": {
    "name": "X-Signature",
    "value": "computed-hmac-signature"
  }
}
```

## Testing Scenarios

### Scenario 1: Main Document Processing Workflow

#### Test Case 1.1: Basic Document Analysis
**Objective**: Validate complete document processing pipeline

**Test Payload**:
```json
{
  "jobId": "test-job-001",
  "dealName": "Test Deal Alpha",
  "documentName": "Financial_Statement_Q3.pdf",
  "documentType": "financial",
  "documentPath": "/deals/test-deal-alpha/documents/Financial_Statement_Q3.pdf",
  "priority": "normal",
  "requestId": "req-test-001",
  "userId": "test-user",
  "timestamp": 1703123456789
}
```

**Expected Results**:
- Document classification: "financial" with confidence > 0.8
- Template discovery: 2+ relevant templates found
- Field mapping: Key fields extracted with confidence scores
- Template population: Data populated with formula preservation
- Result aggregation: Complete results with quality metrics
- Webhook response: Success status with processing summary

**Validation Checklist**:
- [ ] Webhook trigger activates successfully
- [ ] Payload validation passes without errors
- [ ] Document classification returns expected category
- [ ] Template discovery finds relevant templates
- [ ] Field mapping extracts key data points
- [ ] Template population preserves formulas
- [ ] Quality metrics calculated correctly
- [ ] Final webhook response sent to DealDone

#### Test Case 1.2: High Priority Processing
**Test Payload**: Same as 1.1 but with `"priority": "high"`

**Expected Results**:
- Faster processing (priority routing)
- Enhanced validation steps
- Immediate stakeholder notification

#### Test Case 1.3: Low Confidence Document
**Test Payload**: Use unclear document type

**Expected Results**:
- Manual review trigger activated
- Confidence scores below thresholds documented
- Escalation procedures initiated

### Scenario 2: Error Handling Workflow

#### Test Case 2.1: Retryable Error
**Test Payload**:
```json
{
  "jobId": "test-job-002",
  "dealName": "Test Deal Beta",
  "error": {
    "type": "timeout",
    "message": "Request timeout after 30 seconds",
    "stage": "template-population"
  },
  "job": {
    "retryCount": 0,
    "originalPayload": { /* original request */ }
  },
  "failedStage": "template-population"
}
```

**Expected Results**:
- Error classified as "timeout" with high severity
- Retry decision: should retry (true)
- Backoff delay: 5 seconds
- Retry triggered successfully

**Validation Checklist**:
- [ ] Error analysis engine classifies error correctly
- [ ] Retry decision logic works as expected
- [ ] Backoff delay calculated properly
- [ ] Retry preparation engine configures payload
- [ ] Document retry triggered successfully

#### Test Case 2.2: Non-Retryable Error
**Test Payload**: Use authentication error

**Expected Results**:
- Error classified as non-retryable
- Job archived as failed
- Stakeholder notifications sent
- Manual intervention flagged

#### Test Case 2.3: Maximum Retries Exceeded
**Test Payload**: Use job with retryCount = 3

**Expected Results**:
- Retry bypassed
- Final error handling activated
- Comprehensive notifications sent

### Scenario 3: User Corrections Workflow

#### Test Case 3.1: Valid Correction
**Test Payload**:
```json
{
  "userId": "test-user",
  "correction": {
    "stage": "classification",
    "originalValue": "legal",
    "originalConfidence": 0.6,
    "correctedValue": "financial",
    "userConfidence": 1.0,
    "reason": "Document contains financial statements, not legal content"
  },
  "originalData": { /* original processing data */ },
  "documentInfo": {
    "documentId": "doc-001",
    "dealName": "Test Deal",
    "documentType": "mixed"
  }
}
```

**Expected Results**:
- Correction validated as "valid"
- Learning record created and stored
- Confidence models updated
- User feedback generated

**Validation Checklist**:
- [ ] Correction analysis engine processes input
- [ ] Validation router accepts valid correction
- [ ] Learning record stored successfully
- [ ] Confidence models updated
- [ ] Learning completion message generated

#### Test Case 3.2: Invalid Correction
**Test Payload**: Use correction with missing correctedValue

**Expected Results**:
- Validation status: "invalid-missing-value"
- Correction rejected
- User feedback with suggestions provided

### Scenario 4: Cleanup and Maintenance Workflow

#### Test Case 4.1: Scheduled Cleanup
**Trigger**: Cron schedule activation

**Expected Results**:
- Cleanup plan generated with current timestamp
- All cleanup tasks executed in sequence
- Results aggregated and reported
- Cleanup summary sent to DealDone

**Validation Checklist**:
- [ ] Cron trigger activates on schedule
- [ ] Cleanup configuration generated
- [ ] Expired jobs cleanup executes
- [ ] Temp files cleanup executes
- [ ] Cache data cleanup executes
- [ ] Log files cleanup executes
- [ ] Results summary calculated
- [ ] Cleanup report sent successfully

## Debug Procedures

### Common Issues and Solutions

#### Issue 1: Webhook Trigger Not Firing
**Symptoms**: No workflow execution when webhook called

**Debug Steps**:
1. Check webhook URL configuration
2. Verify HTTP method (POST)
3. Validate authentication headers
4. Test with curl command:
```bash
curl -X POST http://n8n-instance/webhook/document-analysis \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"test": "payload"}'
```

#### Issue 2: Node Execution Errors
**Symptoms**: Workflow stops at specific node

**Debug Steps**:
1. Check node configuration parameters
2. Validate JavaScript code syntax
3. Review input/output data format
4. Enable node debug mode
5. Check n8n execution logs

#### Issue 3: API Connection Failures
**Symptoms**: HTTP Request nodes failing

**Debug Steps**:
1. Verify DealDone application is running
2. Check API endpoint URLs
3. Validate authentication credentials
4. Test API endpoints independently
5. Review timeout configurations

#### Issue 4: Data Flow Problems
**Symptoms**: Incorrect data passed between nodes

**Debug Steps**:
1. Check node output using debug mode
2. Validate data transformation logic
3. Review JavaScript code in Code nodes
4. Test with simplified payloads
5. Verify data type conversions

### Debug Tools and Commands

#### Enable n8n Debug Mode
```bash
export N8N_LOG_LEVEL=debug
n8n start
```

#### View Workflow Execution History
```bash
# Access n8n web interface
http://localhost:5678/executions
```

#### Test Individual Nodes
Use n8n's built-in testing feature to execute individual nodes with sample data.

#### Monitor API Calls
```bash
# Monitor DealDone API calls
tail -f /path/to/dealdone/logs/api.log
```

## Performance Testing

### Load Testing Scenarios

#### High Volume Document Processing
**Test Setup**:
- 100 concurrent document analysis requests
- Various document types and sizes
- Monitor processing times and success rates

**Expected Performance**:
- 95% success rate
- Average processing time < 2 minutes
- No memory leaks or resource exhaustion

#### Error Recovery Testing
**Test Setup**:
- Simulate various error conditions
- Monitor retry mechanisms and recovery times
- Validate error notification systems

### Performance Metrics

Track the following metrics during testing:

1. **Workflow Execution Time**
   - Document processing: < 2 minutes
   - Error handling: < 30 seconds
   - User corrections: < 10 seconds
   - Cleanup operations: < 5 minutes

2. **Success Rates**
   - Document processing: > 95%
   - Error recovery: > 90%
   - Cleanup operations: > 99%

3. **Resource Usage**
   - Memory: < 512MB per workflow
   - CPU: < 50% sustained usage
   - Network: < 10MB per document

## Integration Validation

### End-to-End Testing

#### Complete Document Lifecycle
1. Upload document to DealDone
2. Trigger n8n workflow processing
3. Validate results in DealDone
4. Test user corrections workflow
5. Verify error handling scenarios
6. Confirm cleanup operations

#### Multi-Document Batch Processing
1. Upload multiple documents simultaneously
2. Monitor queue management
3. Validate processing order and priorities
4. Check resource usage and performance

### API Integration Tests

#### Webhook Communication
- Test bidirectional webhook communication
- Validate payload formats and authentication
- Monitor webhook response times

#### Data Synchronization
- Verify job status updates
- Validate result storage and retrieval
- Test state consistency between systems

## Production Readiness Checklist

### Security Validation
- [ ] API key authentication working
- [ ] HMAC signature validation active
- [ ] Rate limiting configured
- [ ] HTTPS communication enabled
- [ ] Sensitive data encryption verified

### Reliability Testing
- [ ] Error handling workflows tested
- [ ] Retry mechanisms validated
- [ ] Failover procedures verified
- [ ] Data backup and recovery tested

### Monitoring Setup
- [ ] Workflow execution monitoring
- [ ] Performance metrics collection
- [ ] Error alerting configured
- [ ] Resource usage tracking active

### Documentation Complete
- [ ] Workflow documentation updated
- [ ] API documentation current
- [ ] Troubleshooting guides available
- [ ] Operational runbooks created

## Troubleshooting Reference

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "Authentication failed" | Invalid API key | Update credentials in n8n |
| "Webhook timeout" | Slow API response | Increase timeout values |
| "Invalid payload" | Incorrect data format | Validate JSON schema |
| "Node execution failed" | JavaScript error | Debug code syntax |
| "Connection refused" | DealDone not running | Start DealDone application |

### Support Resources

- **n8n Documentation**: https://docs.n8n.io/
- **DealDone API Reference**: Internal documentation
- **Workflow Configuration**: This repository's n8n-workflows directory
- **Community Support**: n8n community forums

---

**Note**: This testing guide should be executed in a development environment before deploying to production. Always maintain backups of workflow configurations and test data. 