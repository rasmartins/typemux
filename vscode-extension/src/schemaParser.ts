export interface ParsedSchema {
    version: string;
    namespace: string;
    imports: string[];
    types: TypeDefinition[];
    enums: EnumDefinition[];
    unions: UnionDefinition[];
    services: ServiceDefinition[];
}

export interface TypeDefinition {
    name: string;
    documentation: string;
    fields: FieldDefinition[];
    annotations: AnnotationDefinition[];
    lineNumber: number;
}

export interface FieldDefinition {
    name: string;
    type: string;
    fieldNumber?: number;
    required: boolean;
    defaultValue?: string;
    documentation: string;
    annotations: AnnotationDefinition[];
}

export interface EnumDefinition {
    name: string;
    documentation: string;
    values: EnumValue[];
    annotations: AnnotationDefinition[];
    lineNumber: number;
}

export interface EnumValue {
    name: string;
    number: number;
}

export interface UnionDefinition {
    name: string;
    documentation: string;
    options: string[];
    annotations: AnnotationDefinition[];
    lineNumber: number;
}

export interface ServiceDefinition {
    name: string;
    documentation: string;
    methods: MethodDefinition[];
    annotations: AnnotationDefinition[];
    lineNumber: number;
}

export interface MethodDefinition {
    name: string;
    documentation: string;
    request: string;
    response: string;
    annotations: AnnotationDefinition[];
}

export interface AnnotationDefinition {
    name: string;
    value?: string;
}

export class SchemaParser {
    parse(text: string): ParsedSchema {
        const lines = text.split('\n');
        const schema: ParsedSchema = {
            version: '1.0.0',
            namespace: '',
            imports: [],
            types: [],
            enums: [],
            unions: [],
            services: []
        };

        let currentDoc = '';
        let currentAnnotations: AnnotationDefinition[] = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Parse version
            const versionMatch = trimmed.match(/@typemux\("([^"]+)"\)/);
            if (versionMatch) {
                schema.version = versionMatch[1];
                continue;
            }

            // Parse namespace
            const namespaceMatch = trimmed.match(/^namespace\s+([\w.]+)/);
            if (namespaceMatch) {
                schema.namespace = namespaceMatch[1];
                currentDoc = '';
                currentAnnotations = [];
                continue;
            }

            // Parse import
            const importMatch = trimmed.match(/^import\s+"([^"]+)"/);
            if (importMatch) {
                schema.imports.push(importMatch[1]);
                continue;
            }

            // Collect documentation
            if (trimmed.startsWith('///')) {
                currentDoc += trimmed.substring(3).trim() + ' ';
                continue;
            }

            // Collect leading annotations
            if (trimmed.startsWith('@') && !trimmed.includes('(') && !trimmed.includes('{')) {
                const annMatch = trimmed.match(/@([\w.]+)(?:\(([^)]*)\))?/);
                if (annMatch) {
                    currentAnnotations.push({
                        name: '@' + annMatch[1],
                        value: annMatch[2]
                    });
                    continue;
                }
            }

            // Parse type
            const typeMatch = trimmed.match(/^type\s+(\w+)/);
            if (typeMatch && trimmed.includes('{')) {
                const typeDef: TypeDefinition = {
                    name: typeMatch[1],
                    documentation: currentDoc.trim(),
                    fields: [],
                    annotations: [...currentAnnotations],
                    lineNumber: i
                };

                // Parse trailing annotations on type line
                const trailingAnn = this.parseTrailingAnnotations(trimmed);
                typeDef.annotations.push(...trailingAnn);

                // Parse fields
                let j = i + 1;
                let fieldDoc = '';
                let fieldAnnotations: AnnotationDefinition[] = [];

                while (j < lines.length) {
                    const fieldLine = lines[j];
                    const fieldTrimmed = fieldLine.trim();

                    if (fieldTrimmed === '}') break;

                    // Collect field documentation
                    if (fieldTrimmed.startsWith('///')) {
                        fieldDoc += fieldTrimmed.substring(3).trim() + ' ';
                        j++;
                        continue;
                    }

                    // Collect field leading annotations
                    if (fieldTrimmed.startsWith('@') && !fieldTrimmed.includes(':')) {
                        const annMatch = fieldTrimmed.match(/@([\w.]+)(?:\(([^)]*)\))?/);
                        if (annMatch) {
                            fieldAnnotations.push({
                                name: '@' + annMatch[1],
                                value: annMatch[2]
                            });
                        }
                        j++;
                        continue;
                    }

                    // Parse field
                    const fieldMatch = fieldTrimmed.match(/^(\w+)\s*:\s*(\S+)/);
                    if (fieldMatch) {
                        const fieldName = fieldMatch[1];
                        let fieldType = fieldMatch[2];

                        // Extract field number
                        const fieldNumMatch = fieldTrimmed.match(/=\s*(\d+)/);
                        const fieldNumber = fieldNumMatch ? parseInt(fieldNumMatch[1], 10) : undefined;

                        // Check for @required
                        const isRequired = fieldTrimmed.includes('@required');

                        // Extract default value
                        const defaultMatch = fieldTrimmed.match(/@default\(([^)]+)\)/);
                        const defaultValue = defaultMatch ? defaultMatch[1] : undefined;

                        // Parse trailing annotations on field
                        const fieldTrailingAnn = this.parseTrailingAnnotations(fieldTrimmed);

                        // Clean up type (remove field number and annotations)
                        fieldType = fieldType.replace(/\s*=.*$/, '').replace(/@.*$/, '').trim();

                        typeDef.fields.push({
                            name: fieldName,
                            type: fieldType,
                            fieldNumber,
                            required: isRequired,
                            defaultValue,
                            documentation: fieldDoc.trim(),
                            annotations: [...fieldAnnotations, ...fieldTrailingAnn]
                        });

                        fieldDoc = '';
                        fieldAnnotations = [];
                    }

                    j++;
                }

                schema.types.push(typeDef);
                currentDoc = '';
                currentAnnotations = [];
                i = j;
                continue;
            }

            // Parse enum
            const enumMatch = trimmed.match(/^enum\s+(\w+)/);
            if (enumMatch && trimmed.includes('{')) {
                const enumDef: EnumDefinition = {
                    name: enumMatch[1],
                    documentation: currentDoc.trim(),
                    values: [],
                    annotations: [...currentAnnotations],
                    lineNumber: i
                };

                // Parse trailing annotations
                const trailingAnn = this.parseTrailingAnnotations(trimmed);
                enumDef.annotations.push(...trailingAnn);

                // Parse enum values
                let j = i + 1;
                while (j < lines.length) {
                    const valueLine = lines[j];
                    const valueTrimmed = valueLine.trim();

                    if (valueTrimmed === '}') break;

                    if (valueTrimmed.startsWith('//')) {
                        j++;
                        continue;
                    }

                    const valueMatch = valueTrimmed.match(/^(\w+)\s*=\s*(\d+)/);
                    if (valueMatch) {
                        enumDef.values.push({
                            name: valueMatch[1],
                            number: parseInt(valueMatch[2], 10)
                        });
                    }

                    j++;
                }

                schema.enums.push(enumDef);
                currentDoc = '';
                currentAnnotations = [];
                i = j;
                continue;
            }

            // Parse union
            const unionMatch = trimmed.match(/^union\s+(\w+)/);
            if (unionMatch && trimmed.includes('{')) {
                const unionDef: UnionDefinition = {
                    name: unionMatch[1],
                    documentation: currentDoc.trim(),
                    options: [],
                    annotations: [...currentAnnotations],
                    lineNumber: i
                };

                // Parse trailing annotations
                const trailingAnn = this.parseTrailingAnnotations(trimmed);
                unionDef.annotations.push(...trailingAnn);

                // Parse union options
                let j = i + 1;
                while (j < lines.length) {
                    const optionLine = lines[j];
                    const optionTrimmed = optionLine.trim();

                    if (optionTrimmed === '}') break;

                    if (optionTrimmed.startsWith('//')) {
                        j++;
                        continue;
                    }

                    const optionMatch = optionTrimmed.match(/^([\w.]+)/);
                    if (optionMatch) {
                        unionDef.options.push(optionMatch[1]);
                    }

                    j++;
                }

                schema.unions.push(unionDef);
                currentDoc = '';
                currentAnnotations = [];
                i = j;
                continue;
            }

            // Parse service
            const serviceMatch = trimmed.match(/^service\s+(\w+)/);
            if (serviceMatch && trimmed.includes('{')) {
                const serviceDef: ServiceDefinition = {
                    name: serviceMatch[1],
                    documentation: currentDoc.trim(),
                    methods: [],
                    annotations: [...currentAnnotations],
                    lineNumber: i
                };

                // Parse trailing annotations
                const trailingAnn = this.parseTrailingAnnotations(trimmed);
                serviceDef.annotations.push(...trailingAnn);

                // Parse RPC methods
                let j = i + 1;
                let methodDoc = '';
                let methodAnnotations: AnnotationDefinition[] = [];

                while (j < lines.length) {
                    const methodLine = lines[j];
                    const methodTrimmed = methodLine.trim();

                    if (methodTrimmed === '}') break;

                    // Collect method documentation
                    if (methodTrimmed.startsWith('///')) {
                        methodDoc += methodTrimmed.substring(3).trim() + ' ';
                        j++;
                        continue;
                    }

                    // Collect method annotations
                    if (methodTrimmed.startsWith('@')) {
                        const annMatch = methodTrimmed.match(/@([\w.]+)(?:\(([^)]*)\))?/);
                        if (annMatch) {
                            methodAnnotations.push({
                                name: '@' + annMatch[1],
                                value: annMatch[2]
                            });
                        }
                        j++;
                        continue;
                    }

                    // Parse RPC method
                    const rpcMatch = methodTrimmed.match(/^rpc\s+(\w+)\s*\(([^)]+)\)\s*returns\s*\(([^)]+)\)/);
                    if (rpcMatch) {
                        serviceDef.methods.push({
                            name: rpcMatch[1],
                            documentation: methodDoc.trim(),
                            request: rpcMatch[2],
                            response: rpcMatch[3],
                            annotations: [...methodAnnotations]
                        });

                        methodDoc = '';
                        methodAnnotations = [];
                    }

                    j++;
                }

                schema.services.push(serviceDef);
                currentDoc = '';
                currentAnnotations = [];
                i = j;
                continue;
            }

            // Reset doc if we hit a non-comment, non-annotation line
            if (trimmed !== '' && !trimmed.startsWith('//') && !trimmed.startsWith('@')) {
                currentDoc = '';
                currentAnnotations = [];
            }
        }

        return schema;
    }

    private parseTrailingAnnotations(line: string): AnnotationDefinition[] {
        const annotations: AnnotationDefinition[] = [];
        const matches = line.matchAll(/@([\w.]+)(?:\(([^)]*)\))?/g);
        for (const match of matches) {
            annotations.push({
                name: '@' + match[1],
                value: match[2]
            });
        }
        return annotations;
    }
}
