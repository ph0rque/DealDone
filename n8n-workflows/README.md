# DealDone n8n Workflow Integration

This directory contains n8n workflow files for automating document analysis and processing in DealDone.

## Main Workflow: `dealdone-document-processor.json`

### Overview
The DealDone Document Processor is the core n8n workflow that handles end-to-end document analysis for M&A deals. It integrates seamlessly with DealDone's webhook infrastructure to provide automated document classification, data extraction, and template population.

### Workflow Architecture
```
DealDone → Webhook Trigger → Payload Validation → Document Processing → Result Processing → DealDone Response
```

### Key Features
- **Webhook Integration**: Receives triggers from DealDone via secure webhooks
- **Payload Validation**: Ensures all required fields are present and valid
- **Document Processing**: Calls DealDone's AI services for comprehensive analysis
- **Result Aggregation**: Consolidates processing results with statistics
- **Secure Response**: Returns results to DealDone with authentication

## Setup Instructions

### 1. Import Workflow into n8n
1. Open your n8n instance
2. Navigate to **Workflows** → **Import from file**
3. Select `dealdone-document-processor.json`
4. Click **Import workflow**

### 2. Configure Credentials
Set up the following credentials in n8n for DealDone integration:

#### Method 1: Using HTTP Request Node Authentication (Recommended)
For each HTTP Request node in the workflow:

1. **Document Processor Node**:
   - Select the "Document Processor" node
   - In **Authentication** dropdown, select "**Predefined Credential Type**"
   - Choose "**HTTP Header Auth**" from the list
   - Create new credentials with:
     - **Name**: `dealdone-api-key`
     - **Header Name**: `X-API-Key`
     - **Header Value**: Your DealDone API key (generated from DealDone's AuthManager)

2. **Webhook Response Node**:
   - Select the "DealDone Webhook Response" node
   - In **Authentication** dropdown, select "**Predefined Credential Type**"
   - Choose "**HTTP Header Auth**" from the list
   - Create new credentials with:
     - **Name**: `dealdone-webhook-key`
     - **Header Name**: `X-API-Key`
     - **Header Value**: Your DealDone webhook key (generated from DealDone's AuthManager)

#### Method 2: Manual Header Configuration (Current Workflow Setup)
The imported workflow is currently configured for manual header setup:

1. **Document Processor Node**:
   - **Authentication** is set to "**Generic Credential Type**" → "**HTTP Header Auth**"
   - Replace the **X-API-Key** header value `dealdone-api-key` with your actual API key
   - Or change **Authentication** to "**None**" and configure in **Headers** section

2. **Webhook Response Node**:
   - **Authentication** is set to "**Generic Credential Type**" → "**HTTP Header Auth**"
   - Replace the **X-API-Key** header value `dealdone-webhook-key` with your actual webhook key
   - Or change **Authentication** to "**None**" and configure in **Headers** section

> **Note**: The workflow comes pre-configured with placeholder values. You must replace `dealdone-api-key` and `dealdone-webhook-key` with your actual generated keys.

#### Getting DealDone API Keys
Generate the required API keys using DealDone's AuthManager:

1. **API Key for Document Processing**:
   ```javascript
   // In DealDone, call:
   const apiKey = await window.go.main.App.GenerateAPIKey("n8n-integration", ["document:read", "document:analyze"], 365);
   ```

2. **Webhook Key for Result Callbacks**:
   ```javascript
   // In DealDone, call:
   const webhookKey = await window.go.main.App.GenerateWebhookAuthPair("n8n-callback");
   ```

### 3. Configure Webhook URLs
Update the following URLs in the workflow nodes:

#### Document Processor Node
- **URL**: `http://localhost:8081/analyze-document`
- Update to your DealDone instance URL

#### Webhook Response Node
- **URL**: `http://localhost:8080/webhook/results`
- Update to your DealDone webhook endpoint

### 4. Test Configuration (Before Activation)
Before activating the workflow, test your setup:

1. **Test DealDone API Connection**:
   - In n8n, use the **Execute Workflow** button
   - Check the Document Processor node execution
   - Look for authentication errors in the execution log

2. **Verify Webhook Endpoint**:
   - Check that your DealDone webhook server is accessible
   - Test with: `curl -X POST http://localhost:8080/webhook/results -H "Content-Type: application/json" -d "{}"`

### 5. Activate Workflow
1. Click **Active** toggle in the workflow editor
2. The webhook trigger will become available at:
   ```
   https://your-n8n-instance.com/webhook/dealdone-document-analysis
   ```

## Usage

### Triggering the Workflow
Send a POST request to the webhook URL with the following payload:

```json
{
  "dealName": "AcquisitionCorp-TargetInc",
  "filePaths": [
    "/deals/AcquisitionCorp-TargetInc/documents/financial_statements.pdf",
    "/deals/AcquisitionCorp-TargetInc/documents/legal_contracts.pdf"
  ],
  "triggerType": "file_change",
  "jobId": "job-12345-67890",
  "timestamp": 1672531200000,
  "priority": "normal",
  "workflowType": "document-analysis"
}
```

### Expected Response
The workflow will return a structured response:

```json
{
  "jobId": "job-12345-67890",
  "dealName": "AcquisitionCorp-TargetInc",
  "workflowType": "document-analysis",
  "status": "completed",
  "processedDocuments": 2,
  "templatesUpdated": ["Financial_Model.xlsx", "Due_Diligence_Checklist.xlsx"],
  "averageConfidence": 0.87,
  "processingTimeMs": 45000,
  "startTime": 1672531200000,
  "endTime": 1672531245000,
  "results": {
    "documentResults": [...],
    "summary": {...}
  },
  "timestamp": 1672531245000
}
```

## Workflow Nodes Explained

### 1. DealDone Webhook Trigger
- **Type**: Webhook
- **Purpose**: Receives document processing requests from DealDone
- **Path**: `/dealdone-document-analysis`

### 2. Payload Validator
- **Type**: Code (JavaScript)
- **Purpose**: Validates incoming payload structure and required fields
- **Validation**: `dealName`, `filePaths`, `triggerType`, `jobId`, `timestamp`

### 3. Document Processor
- **Type**: HTTP Request
- **Purpose**: Calls DealDone's document analysis API
- **Authentication**: API Key via headers
- **Timeout**: 60 seconds

### 4. Result Processor
- **Type**: Code (JavaScript)
- **Purpose**: Processes analysis results and formats response
- **Output**: Structured result payload for DealDone

### 5. DealDone Webhook Response
- **Type**: HTTP Request
- **Purpose**: Sends processing results back to DealDone
- **Authentication**: API Key + timestamp
- **Endpoint**: DealDone webhook results endpoint

## Integration with DealDone

### Authentication
The workflow uses two-way authentication:
1. **DealDone → n8n**: API key in webhook trigger (optional)
2. **n8n → DealDone**: API key in HTTP requests

### Error Handling
The workflow includes comprehensive error handling:
- Invalid payload validation
- API call failures
- Timeout handling
- Structured error responses

### Monitoring
Monitor workflow execution through:
- n8n execution history
- DealDone job tracking system
- Webhook response logs

## Troubleshooting

### Common Issues

#### 1. Authentication Failures
- **Verify API keys are correctly configured**:
  - Check that placeholder values (`dealdone-api-key`, `dealdone-webhook-key`) are replaced with actual keys
  - Ensure API keys are generated from DealDone's AuthManager
  - Verify key format is correct (should be long alphanumeric strings)
- **Check credential setup**:
  - If using Method 1: Verify credential names match exactly in n8n
  - If using Method 2: Ensure header values are replaced, not credential names
- **Ensure DealDone webhook server is running**:
  - Check that DealDone application is running and webhook server is started
  - Verify webhook endpoints are accessible (test with curl or browser)

#### 2. Connection Timeouts
- Verify DealDone instance is accessible from n8n
- Check firewall and network configurations
- Increase timeout values if processing large documents

#### 3. Payload Validation Errors
- Ensure all required fields are included
- Verify timestamp format (Unix milliseconds)
- Check file paths are absolute and accessible

### Debug Mode
Enable debug mode in n8n:
1. Go to **Settings** → **Log level** → **Debug**
2. Check execution logs for detailed error information
3. Use **Execute Workflow** button for manual testing

## Security Considerations

- Use HTTPS for all webhook communications
- Rotate API keys regularly
- Implement IP whitelisting if possible
- Monitor workflow execution logs
- Use secure credential storage

## Performance Optimization

- Adjust timeout values based on document size
- Consider parallel processing for multiple documents
- Monitor memory usage during large batch processing
- Implement rate limiting if needed

## Future Enhancements

The workflow can be extended with:
- Error handling and retry logic workflows
- User correction processing workflows
- Batch processing optimization
- Real-time progress updates
- Advanced document classification

## Support

For issues or questions:
1. Check n8n execution logs
2. Review DealDone job tracking system
3. Verify webhook endpoint connectivity
4. Consult DealDone API documentation 