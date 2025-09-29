/**
 * FormControlLabel Atom - Atomic Design Pattern
 * 
 * Architecture requirement: "Atomic design pattern with hierarchical component structure" (Section 5.2)
 * Form control with label component
 */

import React from 'react';

export interface FormControlLabelProps {
  control: React.ReactElement;
  label: React.ReactNode;
  labelPlacement?: 'end' | 'start' | 'top' | 'bottom';
  disabled?: boolean;
  className?: string;
}

export const FormControlLabel: React.FC<FormControlLabelProps> = ({
  control,
  label,
  labelPlacement = 'end',
  disabled = false,
  className = '',
  ...props
}) => {
  const placementClasses = {
    end: 'flex items-center',
    start: 'flex items-center flex-row-reverse',
    top: 'flex flex-col items-center',
    bottom: 'flex flex-col-reverse items-center'
  };

  const labelClasses = disabled ? 'text-gray-400' : 'text-gray-700';

  return (
    <label className={`form-control-label ${placementClasses[labelPlacement]} ${disabled ? 'cursor-not-allowed' : 'cursor-pointer'} ${className}`} {...props}>
      {control}
      <span className={`ml-2 text-sm font-medium ${labelClasses}`}>
        {label}
      </span>
    </label>
  );
};
