#!/usr/bin/env node
// Generate annotation reference documentation from annotations.json

const fs = require('fs');
const path = require('path');

// Load annotations.json
const annotationsPath = path.join(__dirname, '..', 'annotations.json');
const annotations = JSON.parse(fs.readFileSync(annotationsPath, 'utf8'));

// Group annotations by category
const schemaAnnotations = annotations.filter(a => a.scope.includes('schema'));
const namespaceAnnotations = annotations.filter(a => a.scope.includes('namespace'));
const typeAnnotations = annotations.filter(a => a.scope.includes('type') || a.scope.includes('enum') || a.scope.includes('union'));
const fieldAnnotations = annotations.filter(a => a.scope.includes('field'));
const methodAnnotations = annotations.filter(a => a.scope.includes('method'));

// Helper to format parameters
function formatParameters(params) {
    if (!params || params.length === 0) return '';

    let result = '\n\n**Parameters:**\n\n';
    for (const param of params) {
        result += `- **${param.name}** (${param.type})${param.required ? ' *required*' : ' *optional*'}: ${param.description}`;
        if (param.validValues && param.validValues.length > 0) {
            result += `\n  - Valid values: \`${param.validValues.join('`, `')}\``;
        }
        if (param.default !== undefined) {
            result += `\n  - Default: \`${JSON.stringify(param.default)}\``;
        }
        result += '\n';
    }
    return result;
}

// Helper to format examples
function formatExamples(examples) {
    if (!examples || examples.length === 0) return '';

    let result = '\n\n**Examples:**\n\n';
    for (const example of examples) {
        result += '```typemux\n' + example + '\n```\n\n';
    }
    return result;
}

// Helper to format formats
function formatFormats(formats) {
    if (!formats || formats.length === 0) return '';
    const formatBadges = formats.map(f => {
        if (f === 'all') return '`all`';
        if (f === 'proto') return '`Protobuf`';
        if (f === 'graphql') return '`GraphQL`';
        if (f === 'openapi') return '`OpenAPI`';
        if (f === 'go') return '`Go`';
        return `\`${f}\``;
    });
    return `**Applies to:** ${formatBadges.join(', ')}`;
}

// Helper to render annotation
function renderAnnotation(annotation) {
    let md = `### ${annotation.name}\n\n`;

    if (annotation.deprecated) {
        md += `> **⚠️ Deprecated:** ${annotation.deprecatedMessage || 'This annotation is deprecated.'}\n\n`;
    }

    md += `${annotation.description}\n\n`;
    md += formatFormats(annotation.formats) + '\n';
    md += formatParameters(annotation.parameters);
    md += formatExamples(annotation.examples);

    return md;
}

// Generate markdown content
let markdown = `---
layout: default
title: Annotation Reference
---

# Annotation Reference

This reference documents all built-in TypeMUX annotations. Annotations provide metadata to customize code generation for different output formats (Protobuf, GraphQL, OpenAPI, Go).

> **📝 Note:** This documentation is automatically generated from [\`annotations.json\`](https://github.com/rasmartins/typemux/blob/main/annotations.json). To see all annotations programmatically:
>
> \`\`\`bash
> typemux annotations
> \`\`\`

---

## Table of Contents

- [Schema-Level Annotations](#schema-level-annotations)
- [Namespace-Level Annotations](#namespace-level-annotations)
- [Type-Level Annotations](#type-level-annotations)
- [Field-Level Annotations](#field-level-annotations)
- [Method-Level Annotations](#method-level-annotations)

---

`;

// Schema annotations
if (schemaAnnotations.length > 0) {
    markdown += `## Schema-Level Annotations\n\n`;
    markdown += `These annotations apply to the entire schema file.\n\n`;
    for (const ann of schemaAnnotations) {
        markdown += renderAnnotation(ann);
    }
    markdown += `---\n\n`;
}

// Namespace annotations
if (namespaceAnnotations.length > 0) {
    markdown += `## Namespace-Level Annotations\n\n`;
    markdown += `These annotations apply to namespace declarations.\n\n`;
    for (const ann of namespaceAnnotations) {
        markdown += renderAnnotation(ann);
    }
    markdown += `---\n\n`;
}

// Type annotations
if (typeAnnotations.length > 0) {
    markdown += `## Type-Level Annotations\n\n`;
    markdown += `These annotations apply to type, enum, and union definitions.\n\n`;
    for (const ann of typeAnnotations) {
        markdown += renderAnnotation(ann);
    }
    markdown += `---\n\n`;
}

// Field annotations
if (fieldAnnotations.length > 0) {
    markdown += `## Field-Level Annotations\n\n`;
    markdown += `These annotations apply to fields within types.\n\n`;
    for (const ann of fieldAnnotations) {
        markdown += renderAnnotation(ann);
    }
    markdown += `---\n\n`;
}

// Method annotations
if (methodAnnotations.length > 0) {
    markdown += `## Method-Level Annotations\n\n`;
    markdown += `These annotations apply to service methods (RPC definitions).\n\n`;
    for (const ann of methodAnnotations) {
        markdown += renderAnnotation(ann);
    }
    markdown += `---\n\n`;
}

// Footer
markdown += `---

## Need More Help?

- See the [Tutorial](tutorial) for practical examples
- Check the [Quick Start](quickstart) guide to get started
- View [Examples](examples) for complete use cases
- Browse the full [Reference](reference) documentation

**Generated from:** [\`annotations.json\`](https://github.com/rasmartins/typemux/blob/main/annotations.json)
**Last updated:** ${new Date().toISOString().split('T')[0]}
`;

// Write to docs directory
const outputPath = path.join(__dirname, '..', 'docs', 'annotations.md');
fs.writeFileSync(outputPath, markdown, 'utf8');

console.log(`✅ Generated annotation documentation: ${outputPath}`);
console.log(`   Total annotations: ${annotations.length}`);
console.log(`   - Schema: ${schemaAnnotations.length}`);
console.log(`   - Namespace: ${namespaceAnnotations.length}`);
console.log(`   - Type: ${typeAnnotations.length}`);
console.log(`   - Field: ${fieldAnnotations.length}`);
console.log(`   - Method: ${methodAnnotations.length}`);
