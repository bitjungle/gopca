import React from 'react';

interface QualityScoreCardProps {
    score: number;
}

export const QualityScoreCard: React.FC<QualityScoreCardProps> = ({ score }) => {
    const getScoreColor = () => {
        if (score >= 80) return 'text-green-600 dark:text-green-400';
        if (score >= 60) return 'text-yellow-600 dark:text-yellow-400';
        return 'text-red-600 dark:text-red-400';
    };

    const getScoreGrade = () => {
        if (score >= 90) return 'Excellent';
        if (score >= 80) return 'Good';
        if (score >= 70) return 'Fair';
        if (score >= 60) return 'Poor';
        return 'Critical';
    };

    const getScoreDescription = () => {
        if (score >= 90) return 'Your data is in excellent condition for PCA analysis.';
        if (score >= 80) return 'Your data is in good condition with minor issues.';
        if (score >= 70) return 'Your data has some quality issues that should be addressed.';
        if (score >= 60) return 'Your data has significant quality issues that will affect analysis.';
        return 'Your data has critical quality issues that must be resolved.';
    };

    return (
        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 border border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-2">
                        Overall Data Quality Score
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                        {getScoreDescription()}
                    </p>
                </div>
                <div className="text-center">
                    <div className={`text-5xl font-bold ${getScoreColor()}`}>
                        {score.toFixed(0)}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                        {getScoreGrade()}
                    </div>
                </div>
            </div>
            
            {/* Score Breakdown Bar */}
            <div className="mt-4">
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                    <div 
                        className={`h-3 rounded-full transition-all duration-500 ${
                            score >= 80 ? 'bg-green-500' :
                            score >= 60 ? 'bg-yellow-500' :
                            'bg-red-500'
                        }`}
                        style={{ width: `${score}%` }}
                    />
                </div>
                <div className="flex justify-between mt-1">
                    <span className="text-xs text-gray-500 dark:text-gray-400">0</span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">50</span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">100</span>
                </div>
            </div>
        </div>
    );
};