import React, { useState } from 'react';
import { X, Folder, Calendar, DollarSign, Building2, Users, MapPin, AlertCircle } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface DealCreationDialogProps {
  onClose: () => void;
  onDealCreated: (dealData: DealFormData) => void;
}

interface DealFormData {
  name: string;
  description: string;
  dealType: 'acquisition' | 'merger' | 'investment' | 'partnership' | 'other';
  industry: string;
  targetCompany: string;
  dealValue: number;
  currency: string;
  expectedCloseDate: string;
  dealStage: 'early' | 'due_diligence' | 'negotiation' | 'closing' | 'closed';
  primaryContact: string;
  location: string;
  priority: 'high' | 'medium' | 'low';
}

export function DealCreationDialog({ onClose, onDealCreated }: DealCreationDialogProps) {
  const [formData, setFormData] = useState<DealFormData>({
    name: '',
    description: '',
    dealType: 'acquisition',
    industry: '',
    targetCompany: '',
    dealValue: 0,
    currency: 'USD',
    expectedCloseDate: '',
    dealStage: 'early',
    primaryContact: '',
    location: '',
    priority: 'medium'
  });

  const [errors, setErrors] = useState<Partial<Record<keyof DealFormData, string>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof DealFormData, string>> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Deal name is required';
    }

    if (!formData.targetCompany.trim()) {
      newErrors.targetCompany = 'Target company is required';
    }

    if (!formData.industry.trim()) {
      newErrors.industry = 'Industry is required';
    }

    if (formData.dealValue <= 0) {
      newErrors.dealValue = 'Deal value must be greater than 0';
    }

    if (!formData.expectedCloseDate) {
      newErrors.expectedCloseDate = 'Expected close date is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);

    try {
      // Call the parent component's handler which will create the actual deal
      await onDealCreated(formData);
      onClose();
    } catch (error) {
      console.error('Error in deal creation dialog:', error);
      toast({
        title: "Error",
        description: "Failed to create deal. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleChange = (field: keyof DealFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: formData.currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 w-full max-w-2xl max-h-[90vh] rounded-lg shadow-xl flex flex-col border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
              <Folder className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Create New Deal</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Set up a new deal and create its folder structure
              </p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

                {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-8">
          {/* Basic Information */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold flex items-center text-gray-900 dark:text-gray-100">
              <Building2 className="h-5 w-5 mr-2 text-blue-600 dark:text-blue-400" />
              Basic Information
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Deal Name <span className="text-red-500">*</span>
                </label>
                 <Input
                   type="text"
                   value={formData.name}
                   onChange={(e) => handleChange('name', e.target.value)}
                   placeholder="e.g., TechCorp Acquisition"
                   className={errors.name ? 'border-red-500' : ''}
                 />
                {errors.name && (
                  <p className="text-red-500 text-xs flex items-center">
                    <AlertCircle className="h-3 w-3 mr-1" />
                    {errors.name}
                  </p>
                )}
              </div>

                            <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Target Company <span className="text-red-500">*</span>
                </label>
                 <Input
                   type="text"
                   value={formData.targetCompany}
                   onChange={(e) => handleChange('targetCompany', e.target.value)}
                   placeholder="e.g., TechCorp Inc."
                   className={errors.targetCompany ? 'border-red-500' : ''}
                 />
                 {errors.targetCompany && (
                   <p className="text-red-500 text-xs flex items-center">
                     <AlertCircle className="h-3 w-3 mr-1" />
                     {errors.targetCompany}
                   </p>
                 )}
               </div>
 
               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                   Deal Type
                 </label>
                 <select
                   value={formData.dealType}
                   onChange={(e) => handleChange('dealType', e.target.value)}
                   className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                 >
                   <option value="acquisition">Acquisition</option>
                   <option value="merger">Merger</option>
                   <option value="investment">Investment</option>
                   <option value="partnership">Partnership</option>
                   <option value="other">Other</option>
                 </select>
               </div>
 
               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                   Industry <span className="text-red-500">*</span>
                 </label>
                 <Input
                   type="text"
                   value={formData.industry}
                   onChange={(e) => handleChange('industry', e.target.value)}
                 placeholder="e.g., Technology, Healthcare"
                 className={errors.industry ? 'border-red-500' : ''}
               />
               {errors.industry && (
                 <p className="text-red-500 text-xs flex items-center">
                   <AlertCircle className="h-3 w-3 mr-1" />
                   {errors.industry}
                 </p>
               )}
             </div>
           </div>

           <div className="space-y-2">
             <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Description</label>
             <textarea
               value={formData.description}
               onChange={(e) => handleChange('description', e.target.value)}
                 placeholder="Brief description of the deal..."
                 rows={3}
                 className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
               />
             </div>
           </div>
 
                      {/* Financial Information */}
           <div className="space-y-4">
             <h3 className="text-lg font-semibold flex items-center text-gray-900 dark:text-gray-100">
               <DollarSign className="h-5 w-5 mr-2 text-blue-600 dark:text-blue-400" />
               Financial Information
             </h3>
             
             <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
               <div className="md:col-span-2 space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                   Deal Value <span className="text-red-500">*</span>
                 </label>
                 <Input
                   type="number"
                   value={formData.dealValue}
                   onChange={(e) => handleChange('dealValue', parseFloat(e.target.value) || 0)}
                   placeholder="1000000"
                   min="0"
                   step="1000"
                   className={errors.dealValue ? 'border-red-500' : ''}
                 />
                 {formData.dealValue > 0 && (
                   <p className="text-xs text-gray-500 dark:text-gray-400">
                     {formatCurrency(formData.dealValue)}
                   </p>
                 )}
                 {errors.dealValue && (
                   <p className="text-red-500 text-xs flex items-center">
                     <AlertCircle className="h-3 w-3 mr-1" />
                     {errors.dealValue}
                   </p>
                 )}
               </div>

               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Currency</label>
                 <select
                   value={formData.currency}
                   onChange={(e) => handleChange('currency', e.target.value)}
                   className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                 >
                   <option value="USD">USD</option>
                   <option value="EUR">EUR</option>
                   <option value="GBP">GBP</option>
                   <option value="JPY">JPY</option>
                   <option value="CAD">CAD</option>
                 </select>
               </div>
             </div>
           </div>
 
                      {/* Timeline & Status */}
           <div className="space-y-4">
             <h3 className="text-lg font-semibold flex items-center text-gray-900 dark:text-gray-100">
               <Calendar className="h-5 w-5 mr-2 text-blue-600 dark:text-blue-400" />
               Timeline & Status
             </h3>
             
             <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                   Expected Close Date <span className="text-red-500">*</span>
                 </label>
                 <Input
                   type="date"
                   value={formData.expectedCloseDate}
                   onChange={(e) => handleChange('expectedCloseDate', e.target.value)}
                   className={errors.expectedCloseDate ? 'border-red-500' : ''}
                 />
                 {errors.expectedCloseDate && (
                   <p className="text-red-500 text-xs flex items-center">
                     <AlertCircle className="h-3 w-3 mr-1" />
                     {errors.expectedCloseDate}
                   </p>
                 )}
               </div>

               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Deal Stage</label>
                 <select
                   value={formData.dealStage}
                   onChange={(e) => handleChange('dealStage', e.target.value)}
                   className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                 >
                   <option value="early">Early Stage</option>
                   <option value="due_diligence">Due Diligence</option>
                   <option value="negotiation">Negotiation</option>
                   <option value="closing">Closing</option>
                   <option value="closed">Closed</option>
                 </select>
               </div>

               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Priority</label>
                 <select
                   value={formData.priority}
                   onChange={(e) => handleChange('priority', e.target.value)}
                   className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                 >
                   <option value="high">High Priority</option>
                   <option value="medium">Medium Priority</option>
                   <option value="low">Low Priority</option>
                 </select>
               </div>
             </div>
           </div>
 
                      {/* Additional Information */}
           <div className="space-y-4">
             <h3 className="text-lg font-semibold flex items-center text-gray-900 dark:text-gray-100">
               <Users className="h-5 w-5 mr-2 text-blue-600 dark:text-blue-400" />
               Additional Information
             </h3>
             
             <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Primary Contact</label>
                 <Input
                   type="text"
                   value={formData.primaryContact}
                   onChange={(e) => handleChange('primaryContact', e.target.value)}
                   placeholder="e.g., John Smith, CEO"
                 />
               </div>

               <div className="space-y-2">
                 <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Location</label>
                 <Input
                   type="text"
                   value={formData.location}
                   onChange={(e) => handleChange('location', e.target.value)}
                   placeholder="e.g., San Francisco, CA"
                 />
               </div>
             </div>
           </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <div className="text-sm text-gray-500 dark:text-gray-400">
            <span className="text-red-500">*</span> Required fields
          </div>
          <div className="flex space-x-3">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button 
              type="submit" 
              onClick={handleSubmit}
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Creating...
                </>
              ) : (
                <>
                  <Folder className="h-4 w-4 mr-2" />
                  Create Deal
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
} 