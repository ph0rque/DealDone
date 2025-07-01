import { toast } from '../hooks/use-toast'

export interface AppError {
  code: string
  message: string
  details?: string
  severity: 'low' | 'medium' | 'high' | 'critical'
}

export class ErrorService {
  private static errorMessages: Record<string, string> = {
    // File operation errors
    'ENOENT': 'File or folder not found',
    'EACCES': 'Permission denied',
    'EEXIST': 'File or folder already exists',
    'ENOSPC': 'Not enough space on disk',
    'EMFILE': 'Too many open files',
    'ENOTDIR': 'Not a directory',
    'EISDIR': 'Is a directory',
    'ENOTEMPTY': 'Directory not empty',
    
    // Network errors
    'NETWORK_ERROR': 'Network connection failed',
    'TIMEOUT': 'Operation timed out',
    'SERVER_ERROR': 'Server error occurred',
    
    // App-specific errors
    'INVALID_PATH': 'Invalid file or folder path',
    'OPERATION_CANCELLED': 'Operation was cancelled',
    'UNSUPPORTED_FORMAT': 'Unsupported file format',
    'FILE_TOO_LARGE': 'File is too large',
    
    // Generic errors
    'UNKNOWN_ERROR': 'An unexpected error occurred'
  }

  public static getErrorCode(error: Error | string): string {
    if (typeof error === 'string') {
      return 'UNKNOWN_ERROR'
    }

    // Check for common file system error codes
    if ('code' in error && typeof error.code === 'string') {
      return error.code
    }

    // Check error message for known patterns
    const message = error.message.toLowerCase()
    if (message.includes('permission')) return 'EACCES'
    if (message.includes('not found')) return 'ENOENT'
    if (message.includes('already exists')) return 'EEXIST'
    if (message.includes('network')) return 'NETWORK_ERROR'
    if (message.includes('timeout')) return 'TIMEOUT'

    return 'UNKNOWN_ERROR'
  }

  public static getUserFriendlyMessage(error: Error | string): string {
    const code = this.getErrorCode(error)
    return this.errorMessages[code] || this.errorMessages['UNKNOWN_ERROR']
  }

  public static getSeverity(error: Error | string): AppError['severity'] {
    const code = this.getErrorCode(error)
    
    switch (code) {
      case 'EACCES':
      case 'ENOSPC':
        return 'high'
      case 'ENOENT':
      case 'EEXIST':
      case 'NETWORK_ERROR':
        return 'medium'
      case 'OPERATION_CANCELLED':
        return 'low'
      default:
        return 'medium'
    }
  }

  public static createAppError(error: Error | string, details?: string): AppError {
    const code = this.getErrorCode(error)
    const message = this.getUserFriendlyMessage(error)
    const severity = this.getSeverity(error)

    return {
      code,
      message,
      details: details || (typeof error === 'object' ? error.message : error),
      severity
    }
  }

  public static handleError(error: Error | string, options?: {
    showToast?: boolean
    title?: string
    details?: string
  }): AppError {
    const appError = this.createAppError(error, options?.details)
    
    // Log error for debugging
    console.error('Error handled by ErrorService:', {
      code: appError.code,
      message: appError.message,
      details: appError.details,
      severity: appError.severity,
      originalError: error
    })

    // Show toast notification if requested
    if (options?.showToast !== false) {
      const variant = appError.severity === 'high' || appError.severity === 'critical' 
        ? 'destructive' 
        : 'default'

      toast({
        title: options?.title || 'Error',
        description: appError.message,
        variant
      })
    }

    return appError
  }

  public static handleFileOperationError(
    operation: string,
    error: Error | string,
    filename?: string
  ): AppError {
    const appError = this.createAppError(error)
    const fileContext = filename ? ` for "${filename}"` : ''
    
    const operationMessages: Record<string, string> = {
      create: `Failed to create file${fileContext}`,
      copy: `Failed to copy file${fileContext}`,
      move: `Failed to move file${fileContext}`,
      delete: `Failed to delete file${fileContext}`,
      rename: `Failed to rename file${fileContext}`,
      open: `Failed to open file${fileContext}`,
      search: 'Failed to search files'
    }

    const title = operationMessages[operation] || `File operation failed${fileContext}`

    return this.handleError(error, {
      title,
      showToast: true
    })
  }

  public static isRetryableError(error: AppError): boolean {
    const retryableCodes = ['NETWORK_ERROR', 'TIMEOUT', 'SERVER_ERROR']
    return retryableCodes.includes(error.code)
  }

  public static shouldShowDetails(error: AppError): boolean {
    return error.severity === 'high' || error.severity === 'critical'
  }
} 