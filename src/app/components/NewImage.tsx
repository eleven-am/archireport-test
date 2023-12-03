'use client';

import {memo, useState} from "react";

interface NewImageProps {
    src: string;
    alt: string;
    darkSrc: string;
    className?: string;
    loading?: 'eager' | 'lazy';
}

export const NewImage = memo(function NewImage({src, alt, darkSrc, className, loading}: NewImageProps) {
    const [isDark] = useState(true);

    return (
        <img
            src={isDark ? darkSrc : src}
            alt={alt}
            loading={loading}
            className={className}
        />
    );
});
