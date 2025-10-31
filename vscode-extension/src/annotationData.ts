import * as fs from 'fs';
import * as path from 'path';

export interface AnnotationParameter {
    name: string;
    type: string;
    required: boolean;
    description: string;
    validValues?: string[];
    default?: any;
}

export interface AnnotationMetadata {
    name: string;
    scope: string[];
    formats: string[];
    parameters?: AnnotationParameter[];
    description: string;
    examples?: string[];
    deprecated?: boolean;
    deprecatedMessage?: string;
}

class AnnotationRegistry {
    private annotations: Map<string, AnnotationMetadata> = new Map();

    constructor() {
        this.loadAnnotations();
    }

    private loadAnnotations() {
        try {
            const annotationsPath = path.join(__dirname, '..', 'annotations.json');
            const data = fs.readFileSync(annotationsPath, 'utf8');
            const annotationsList: AnnotationMetadata[] = JSON.parse(data);

            for (const annotation of annotationsList) {
                this.annotations.set(annotation.name, annotation);
            }
        } catch (error) {
            console.error('Failed to load annotations.json:', error);
        }
    }

    getAnnotation(name: string): AnnotationMetadata | undefined {
        return this.annotations.get(name);
    }

    getAllAnnotations(): AnnotationMetadata[] {
        return Array.from(this.annotations.values());
    }

    getAnnotationsByScope(scope: string): AnnotationMetadata[] {
        return this.getAllAnnotations().filter(a => a.scope.includes(scope));
    }

    getAnnotationsByFormat(format: string): AnnotationMetadata[] {
        return this.getAllAnnotations().filter(a =>
            a.formats.includes(format) || a.formats.includes('all')
        );
    }
}

// Singleton instance
export const annotationRegistry = new AnnotationRegistry();
