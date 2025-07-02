# DealDone n8n Webhook Trigger Setup Guide

This guide provides detailed instructions for configuring webhook triggers in n8n to receive requests from DealDone.

## Overview

DealDone supports multiple webhook trigger types for different use cases:
- **Document Analysis**: Primary workflow for document processing
- **Batch Processing**: Handling multiple documents simultaneously  
- **Error Handling**: Managing failed workflows and retries
- **User Corrections**: Processing user feedback and learning
- **Health Check**: System monitoring and status verification

## Core Webhook Trigger Configuration

### 1. Document Analysis Trigger (Primary)

**Purpose**: Receives document analysis requests from DealDone frontend

#### Basic Configuration
```json
{
  "httpMethod": "POST",
  "path": "dealdone-document-analysis",
  "authenticationMethod": "queryAuth",
  "responseMode": "responseNode"
}
```

#### Enhanced Configuration
```json
{
  "httpMethod": "POST", 
  "path": "dealdone-document-analysis",
  "authenticationMethod": "queryAuth",
  "queryAuth": {
    "name": "api_key",
    "value": "={{ $credentials.dealDoneApiKey }}"
  },
  "responseMode": "responseNode",
  "options": {
    "allowedOrigins": "http://localhost:8080,https://dealdone.local",
    "httpCodeOnSuccess": 200,
    "rawBody": false
  }
}
```

#### Required Headers
- `Content-Type`: application/json
- `X-DealDone-Version`: Version compatibility (e.g., "1.1.0")
- `X-Request-ID`: Unique request identifier

#### Optional Headers
- `X-Client-Info`: Client information
- `X-Priority`: Processing priority (1=low, 2=normal, 3=high)
- `X-Timeout`: Custom timeout in seconds

#### Sample Payload
```json
{
  "dealName": "AcquisitionCorp-TargetInc",
  "filePaths": [
    "/deals/AcquisitionCorp-TargetInc/documents/financial_statements.pdf"
  ],
  "triggerType": "user_button",
  "workflowType": "document-analysis",
  "jobId": "job-12345-67890",
  "priority": 2,
  "timestamp": 1672531200000,
  "metadata": {
    "source": "frontend",
    "userAgent": "DealDone/1.1.0"
  }
}
```

### 2. Batch Processing Trigger

**Purpose**: Handles multiple documents in a single workflow execution

#### Configuration
```json
{
  "httpMethod": "POST",
  "path": "dealdone-batch-processing", 
  "authenticationMethod": "queryAuth",
  "options": {
    "allowedOrigins": "http://localhost:8080,https://dealdone.local",
    "maxFileSize": "500MB",
    "timeout": 900000
  }
}
```

#### Required Headers
- `Content-Type`: application/json
- `X-DealDone-Version`: Version compatibility
- `X-Request-ID`: Unique request identifier
- `X-Batch-Size`: Number of items in batch

#### Sample Payload
```json
{
  "batchId": "batch-12345-67890",
  "dealName": "AcquisitionCorp-TargetInc",
  "batchType": "deal_analysis",
  "items": [
    {
      "itemId": "item1",
      "itemType": "document", 
      "itemPath": "/documents/doc1.pdf"
    }
  ],
  "priority": 2,
  "timestamp": 1672531200000
}
```

### 3. Error Handling Trigger

**Purpose**: Processes error scenarios and retry workflows

#### Configuration
```json
{
  "httpMethod": "POST",
  "path": "dealdone-error-handling",
  "authenticationMethod": "queryAuth",
  "options": {
    "allowedOrigins": "http://localhost:8080,https://dealdone.local",
    "maxFileSize": "10MB",
    "timeout": 120000
  }
}
```

#### Required Headers
- `Content-Type`: application/json
- `X-DealDone-Version`: Version compatibility
- `X-Request-ID`: Unique request identifier
- `X-Error-Type`: Type of error being processed

#### Sample Payload
```json
{
  "originalJobId": "job-12345-67890",
  "errorJobId": "error-12345-67890",
  "dealName": "AcquisitionCorp-TargetInc",
  "errorType": "processing_timeout",
  "retryAttempt": 1,
  "maxRetries": 3,
  "retryStrategy": "exponential",
  "recoveryAction": "retry",
  "timestamp": 1672531200000
}
```

### 4. User Corrections Trigger

**Purpose**: Handles user feedback and learning improvements

#### Configuration
```json
{
  "httpMethod": "POST",
  "path": "dealdone-user-corrections",
  "authenticationMethod": "queryAuth", 
  "options": {
    "allowedOrigins": "http://localhost:8080,https://dealdone.local",
    "maxFileSize": "5MB",
    "timeout": 180000
  }
}
```

#### Required Headers
- `Content-Type`: application/json
- `X-DealDone-Version`: Version compatibility
- `X-Request-ID`: Unique request identifier
- `X-Correction-Type`: Type of correction (manual, automated)

#### Sample Payload
```json
{
  "correctionId": "corr-12345-67890",
  "originalJobId": "job-12345-67890",
  "dealName": "AcquisitionCorp-TargetInc",
  "templatePath": "/templates/deal_template.xlsx",
  "corrections": [
    {
      "fieldName": "dealValue",
      "originalValue": "1000000",
      "correctedValue": "1200000",
      "originalConfidence": 0.75,
      "userConfidence": 0.95,
      "correctionReason": "wrong_extraction"
    }
  ],
  "correctionType": "manual",
  "applyToSimilar": true,
  "confidence": 0.95,
  "timestamp": 1672531200000
}
```

### 5. Health Check Trigger

**Purpose**: System monitoring and status verification

#### Configuration
```json
{
  "httpMethod": "GET",
  "path": "dealdone-health-check",
  "authenticationMethod": "queryAuth",
  "options": {
    "allowedOrigins": "*",
    "maxFileSize": "1MB",
    "timeout": 30000
  }
}
```

#### Sample Payload
```json
{
  "checkId": "health-12345",
  "checkType": "system",
  "components": ["database", "ai_service", "file_system"],
  "timestamp": 1672531200000
}
```

## Authentication Configuration

### API Key Authentication (Recommended)

1. **Setup in n8n**:
   - Authentication Method: `queryAuth`
   - Parameter Name: `api_key`
   - Value: Use DealDone generated API key

2. **DealDone Configuration**:
   ```javascript
   // Generate API key in DealDone
   const apiKey = "Q01qQqDgCaEwMchSgCBf1PNeW1CFnInQ3_chMgWzN7A=";
   ```

3. **URL Format**:
   ```
   https://your-n8n-instance.com/webhook/dealdone-document-analysis?api_key=YOUR_API_KEY
   ```

### HMAC Signature Authentication (Advanced)

1. **Header Configuration**:
   - Header Name: `X-Signature`
   - Algorithm: HMAC-SHA256
   - Secret: DealDone HMAC secret

2. **Signature Generation**:
   ```javascript
   const signature = crypto
     .createHmac('sha256', 'YOUR_HMAC_SECRET')
     .update(JSON.stringify(payload))
     .digest('hex');
   ```

## Advanced Trigger Features

### Priority-Based Processing

Configure different processing paths based on request priority:

```javascript
// Priority Router Node
const priority = $json.metadata?.priority || 'normal';
switch(priority) {
  case 'high':
    return [{json: {...$json, processingPath: 'high-priority'}}];
  case 'low':
    return [{json: {...$json, processingPath: 'low-priority'}}];
  default:
    return [{json: {...$json, processingPath: 'normal-priority'}}];
}
```

### Enhanced Validation

Implement comprehensive payload validation:

```javascript
// Enhanced Validation Node
const payload = $json;
const headers = $node['Webhook Trigger'].json.headers;

// Header validation
const requiredHeaders = ['content-type', 'x-dealdone-version', 'x-request-id'];
const missingHeaders = requiredHeaders.filter(h => !headers[h]);
if (missingHeaders.length > 0) {
  throw new Error(`Missing headers: ${missingHeaders.join(', ')}`);
}

// Payload validation
const requiredFields = ['dealName', 'filePaths', 'jobId', 'timestamp'];
const missingFields = requiredFields.filter(f => !payload[f]);
if (missingFields.length > 0) {
  throw new Error(`Missing fields: ${missingFields.join(', ')}`);
}

// Timestamp validation (within 1 hour)
const age = Date.now() - payload.timestamp;
if (age > 3600000) {
  throw new Error(`Payload too old: ${age}ms`);
}

return payload;
```

### Error Handling Enhancement

Implement detailed error analysis:

```javascript
// Error Handler Node
const error = $json.error || {};
const originalPayload = $json;

const errorResponse = {
  jobId: originalPayload.jobId,
  status: 'failed',
  error: {
    type: error.name || 'UNKNOWN_ERROR',
    message: error.message || 'Unknown error',
    timestamp: Date.now(),
    recoverable: !['ValidationError', 'AuthError'].includes(error.name),
    retryDelay: error.name === 'TimeoutError' ? 60000 : 30000
  },
  metadata: originalPayload.metadata
};

return errorResponse;
```

## Rate Limiting Configuration

Configure rate limits per trigger type:

| Trigger Type | Rate Limit | Burst Limit |
|--------------|------------|-------------|
| Document Analysis | 60/minute | 10/second |
| Batch Processing | 10/minute | 2/second |
| Error Handling | 30/minute | 5/second |
| User Corrections | 20/minute | 3/second |
| Health Check | 120/minute | 20/second |

## Monitoring and Logging

### Essential Metrics to Track

1. **Request Metrics**:
   - Request count per trigger type
   - Response times
   - Error rates
   - Payload sizes

2. **Processing Metrics**:
   - Queue depths
   - Processing times
   - Success/failure rates
   - Resource utilization

3. **Business Metrics**:
   - Documents processed
   - Templates updated
   - User corrections applied
   - Average confidence scores

### Logging Configuration

```json
{
  "logging": {
    "enabled": true,
    "level": "info",
    "include_payload": false,
    "include_headers": true,
    "retention_days": 30
  }
}
```

## Testing and Validation

### Test Commands

1. **Document Analysis Test**:
   ```bash
   curl -X POST 'https://your-n8n.com/webhook/dealdone-document-analysis?api_key=YOUR_KEY' \
     -H 'Content-Type: application/json' \
     -H 'X-DealDone-Version: 1.1.0' \
     -H 'X-Request-ID: test-123' \
     -d @sample_payload.json
   ```

2. **Health Check Test**:
   ```bash
   curl -X GET 'https://your-n8n.com/webhook/dealdone-health-check' \
     -H 'X-DealDone-Version: 1.1.0'
   ```

### Validation Checklist

- [ ] Webhook URLs are accessible from DealDone
- [ ] Authentication is properly configured
- [ ] Payload validation works correctly
- [ ] Error handling returns proper responses
- [ ] Rate limiting is enforced
- [ ] Logging captures essential information
- [ ] Monitoring dashboards show metrics

## Troubleshooting

### Common Issues

1. **Authentication Failures**:
   - Verify API key is correctly configured
   - Check query parameter name matches ("api_key")
   - Ensure key hasn't expired

2. **Payload Validation Errors**:
   - Verify required fields are present
   - Check timestamp is recent (< 1 hour)
   - Validate JSON structure

3. **Timeout Issues**:
   - Increase timeout values for large payloads
   - Check network connectivity
   - Monitor resource usage

4. **Rate Limit Exceeded**:
   - Implement exponential backoff
   - Reduce request frequency
   - Use batch processing for multiple items

### Debug Mode

Enable debug logging to troubleshoot issues:

```json
{
  "debug": {
    "enabled": true,
    "log_payloads": true,
    "log_headers": true,
    "log_responses": true
  }
}
```

## Security Best Practices

1. **API Key Management**:
   - Rotate keys regularly (every 90 days)
   - Use different keys for different environments
   - Monitor key usage and detect anomalies

2. **Network Security**:
   - Use HTTPS for all communications
   - Implement IP whitelisting if possible
   - Monitor for suspicious activity

3. **Payload Security**:
   - Validate all input data
   - Sanitize file paths
   - Limit payload sizes
   - Implement request signatures (HMAC)

4. **Error Handling**:
   - Don't expose sensitive information in errors
   - Log security events
   - Implement proper error responses

This comprehensive webhook trigger configuration ensures robust, secure, and scalable communication between DealDone and n8n workflows. 