import * as vscode from 'vscode';

export class TypeMuxHoverProvider implements vscode.HoverProvider {
    private annotations: Map<string, { description: string; usage: string; example: string }>;
    private builtinTypes: Map<string, { description: string; usage: string; generators: string }>;
    private outputChannel: vscode.OutputChannel;

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
        this.outputChannel.appendLine('TypeMuxHoverProvider constructor called');

        // Builtin type documentation
        this.builtinTypes = new Map([
            ['string', {
                description: 'Unicode text string',
                usage: 'Used for text data of any length',
                generators: '**Protobuf:** `string` | **GraphQL:** `String` | **OpenAPI:** `string`'
            }],
            ['int32', {
                description: '32-bit signed integer',
                usage: 'Range: -2,147,483,648 to 2,147,483,647',
                generators: '**Protobuf:** `int32` | **GraphQL:** `Int` | **OpenAPI:** `integer (int32)`'
            }],
            ['int64', {
                description: '64-bit signed integer',
                usage: 'Range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807',
                generators: '**Protobuf:** `int64` | **GraphQL:** `Int` | **OpenAPI:** `integer (int64)`'
            }],
            ['float32', {
                description: '32-bit floating point number',
                usage: 'Single precision floating point',
                generators: '**Protobuf:** `float` | **GraphQL:** `Float` | **OpenAPI:** `number (float)`'
            }],
            ['float64', {
                description: '64-bit floating point number',
                usage: 'Double precision floating point',
                generators: '**Protobuf:** `double` | **GraphQL:** `Float` | **OpenAPI:** `number (double)`'
            }],
            ['bool', {
                description: 'Boolean true/false value',
                usage: 'Represents true or false',
                generators: '**Protobuf:** `bool` | **GraphQL:** `Boolean` | **OpenAPI:** `boolean`'
            }],
            ['bytes', {
                description: 'Binary data',
                usage: 'Used for raw binary data, files, or blobs',
                generators: '**Protobuf:** `bytes` | **GraphQL:** `String` (base64) | **OpenAPI:** `string (byte)`'
            }],
            ['timestamp', {
                description: 'Date and time',
                usage: 'Represents a point in time',
                generators: '**Protobuf:** `google.protobuf.Timestamp` | **GraphQL:** `String` (ISO 8601) | **OpenAPI:** `string (date-time)`'
            }],
            ['map', {
                description: 'Key-value map',
                usage: 'Dictionary or associative array with string keys',
                generators: '**Protobuf:** `map<string, T>` | **GraphQL:** `JSON` scalar | **OpenAPI:** `object (additionalProperties)`'
            }]
        ]);

        this.annotations = new Map([
            ['@required', {
                description: 'Marks a field as required/non-nullable',
                usage: 'Use on type fields to indicate they must have a value',
                example: '```typemux\ntype User {\n    id: string @required\n    name: string @required\n}\n```'
            }],
            ['@default', {
                description: 'Specifies a default value for a field',
                usage: 'Use with a value in parentheses to set a default',
                example: '```typemux\ntype Config {\n    enabled: bool @default(true)\n    maxRetries: int32 @default(3)\n    message: string @default("Hello")\n}\n```'
            }],
            ['@exclude', {
                description: 'Excludes a field from specific generators',
                usage: 'Specify one or more generators to exclude this field from',
                example: '```typemux\ntype User {\n    id: string @required\n    internalData: string @exclude(openapi)\n    debugInfo: string @exclude(proto, graphql)\n}\n```'
            }],
            ['@only', {
                description: 'Includes a field only in specific generators',
                usage: 'Specify which generators should include this field',
                example: '```typemux\ntype User {\n    id: string @required\n    restOnlyField: string @only(openapi)\n    protoOnlyField: bytes @only(proto)\n}\n```'
            }],
            ['@http', {
                description: 'Specifies the HTTP method for an RPC endpoint',
                usage: 'Use on service methods to define REST API behavior',
                example: '```typemux\nservice UserService {\n    rpc GetUser(Request) returns (Response)\n        @http.method(GET)\n        @http.path("/api/users/{id}")\n    \n    rpc CreateUser(Request) returns (Response)\n        @http.method(POST)\n        @http.path("/api/users")\n}\n```'
            }],
            ['@graphql', {
                description: 'Specifies the GraphQL operation type',
                usage: 'Use query, mutation, or subscription',
                example: '```typemux\nservice UserService {\n    rpc GetUser(Request) returns (Response)\n        @graphql(query)\n    \n    rpc CreateUser(Request) returns (Response)\n        @graphql(mutation)\n    \n    rpc SubscribeToUpdates(Request) returns (Response)\n        @graphql(subscription)\n}\n```'
            }],
            ['@proto.option', {
                description: 'Adds Protobuf-specific field options',
                usage: 'Use with Protobuf options in brackets',
                example: '```typemux\ntype User {\n    tags: []string @proto.option([packed = false])\n    metadata: bytes @proto.option([retention = RETENTION_SOURCE])\n}\n```'
            }],
            ['@graphql.directive', {
                description: 'Adds GraphQL directives to types or fields',
                usage: 'Use for GraphQL Federation, authorization, etc.',
                example: '```typemux\ntype User @graphql.directive(@key(fields: "id")) {\n    id: string @required @graphql.directive(@external)\n    email: string @required\n}\n```'
            }],
            ['@openapi.extension', {
                description: 'Adds OpenAPI extensions (x-* properties)',
                usage: 'Use JSON object with extension properties',
                example: '```typemux\ntype Product @openapi.extension({"x-internal-id": "prod-v1"}) {\n    price: float64 @openapi.extension({"x-format": "currency"})\n}\n\n// Nested extensions\ntype Config @openapi.extension({"x-metadata": {"version": "v2", "features": ["auth"]}}) {\n    timeout: int32 @openapi.extension({"x-validation": {"min": 1, "max": 3600}})\n}\n```'
            }],
            ['@proto.name', {
                description: 'Specifies custom Protobuf message/type name',
                usage: 'Use leading or trailing annotation to override the default name',
                example: '```typemux\n// Leading annotation\n@proto.name("UserV2")\ntype User {\n    id: string @required\n}\n\n// Trailing annotation\ntype Product @proto.name("ProductV3") {\n    id: string @required\n}\n```'
            }],
            ['@graphql.name', {
                description: 'Specifies custom GraphQL type name',
                usage: 'Use leading or trailing annotation to override the default name',
                example: '```typemux\n@graphql.name("UserAccount")\ntype User {\n    id: string @required\n}\n```'
            }],
            ['@openapi.name', {
                description: 'Specifies custom OpenAPI schema name',
                usage: 'Use leading or trailing annotation to override the default name',
                example: '```typemux\n@openapi.name("UserProfile")\ntype User {\n    id: string @required\n}\n```'
            }],
            ['@success', {
                description: 'Specifies additional HTTP success status codes',
                usage: 'Use comma-separated status codes (201, 202, 204, etc.)',
                example: '```typemux\nservice UserService {\n    rpc CreateUser(Request) returns (Response)\n        @http.method(POST)\n        @http.path("/api/users")\n        @http.success(201)\n        @http.errors(400,409,500)\n    \n    rpc DeleteUser(Request) returns (Response)\n        @http.method(DELETE)\n        @http.path("/api/users/{id}")\n        @http.success(204)\n        @http.errors(404,500)\n}\n```'
            }],
            ['@errors', {
                description: 'Specifies expected HTTP error status codes',
                usage: 'Use comma-separated status codes (400, 404, 500, etc.)',
                example: '```typemux\nservice UserService {\n    rpc GetUser(Request) returns (Response)\n        @http.method(GET)\n        @http.path("/api/users/{id}")\n        @http.errors(404,500)\n    \n    rpc CreateUser(Request) returns (Response)\n        @http.method(POST)\n        @http.path("/api/users")\n        @http.success(201)\n        @http.errors(400,409,422,500)\n}\n```'
            }]
        ]);
    }

    provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<vscode.Hover> {
        this.outputChannel.appendLine(`=== provideHover called at ${position.line}:${position.character} ===`);

        const line = document.lineAt(position.line);
        const lineText = line.text;
        this.outputChannel.appendLine(`Line text: "${lineText}"`);

        // Check for annotations first (including dot notation like @proto.option)
        const annotationMatch = lineText.match(/@(\w+(?:\.\w+)?)/g);
        if (annotationMatch) {
            this.outputChannel.appendLine(`Found annotations: ${annotationMatch.join(', ')}`);
            // Find which annotation the cursor is on
            for (const match of annotationMatch) {
                const index = lineText.indexOf(match);
                const endIndex = index + match.length;

                if (position.character >= index && position.character <= endIndex) {
                    this.outputChannel.appendLine(`Cursor is on annotation: ${match}`);
                    const annotationInfo = this.annotations.get(match);
                    if (annotationInfo) {
                        this.outputChannel.appendLine(`Returning hover for annotation: ${match}`);
                        const markdown = new vscode.MarkdownString();
                        markdown.appendMarkdown(`### ${match}\n\n`);
                        markdown.appendMarkdown(`${annotationInfo.description}\n\n`);
                        markdown.appendMarkdown(`**Usage:** ${annotationInfo.usage}\n\n`);
                        markdown.appendMarkdown(`**Example:**\n\n${annotationInfo.example}`);
                        markdown.isTrusted = true;

                        return new vscode.Hover(markdown);
                    }
                }
            }
        }

        // Check for type names (including numbers like int32, int64, float32, etc.)
        const typePattern = /\b[a-zA-Z_][a-zA-Z0-9_]*\b/;
        const wordRange = document.getWordRangeAtPosition(position, typePattern);
        this.outputChannel.appendLine(`Word range: ${wordRange ? `${wordRange.start.character}-${wordRange.end.character}` : 'null'}`);
        if (wordRange) {
            const word = document.getText(wordRange);
            this.outputChannel.appendLine(`Word at cursor: "${word}"`);

            // Check for builtin types
            const builtinInfo = this.builtinTypes.get(word);
            this.outputChannel.appendLine(`Checking builtin type "${word}": ${builtinInfo ? 'FOUND' : 'not found'}`);
            if (builtinInfo) {
                this.outputChannel.appendLine(`Returning hover for builtin type: ${word}`);
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`### \`${word}\` (builtin type)\n\n`);
                markdown.appendMarkdown(`${builtinInfo.description}\n\n`);
                markdown.appendMarkdown(`**${builtinInfo.usage}**\n\n`);
                markdown.appendMarkdown(`**Generated as:**\n\n${builtinInfo.generators}`);
                markdown.isTrusted = true;

                return new vscode.Hover(markdown);
            }

            // Check for custom types
            const customTypeInfo = this.findCustomType(document, word);
            this.outputChannel.appendLine(`Checking custom type "${word}": ${customTypeInfo ? 'FOUND' : 'not found'}`);
            if (customTypeInfo) {
                this.outputChannel.appendLine(`Returning hover for custom type: ${word} (${customTypeInfo.kind})`);
                const markdown = new vscode.MarkdownString();
                markdown.appendMarkdown(`### \`${word}\` (${customTypeInfo.kind})\n\n`);

                if (customTypeInfo.doc) {
                    markdown.appendMarkdown(`${customTypeInfo.doc}\n\n`);
                }

                markdown.appendMarkdown(`**Defined at:** line ${customTypeInfo.line + 1}\n\n`);

                if (customTypeInfo.fields && customTypeInfo.fields.length > 0) {
                    markdown.appendMarkdown(`**Fields:**\n`);
                    for (const field of customTypeInfo.fields) {
                        markdown.appendMarkdown(`- \`${field}\`\n`);
                    }
                } else if (customTypeInfo.values && customTypeInfo.values.length > 0) {
                    markdown.appendMarkdown(`**Values:**\n`);
                    for (const value of customTypeInfo.values) {
                        markdown.appendMarkdown(`- \`${value}\`\n`);
                    }
                } else if (customTypeInfo.options && customTypeInfo.options.length > 0) {
                    markdown.appendMarkdown(`**Options:**\n`);
                    for (const option of customTypeInfo.options) {
                        markdown.appendMarkdown(`- \`${option}\`\n`);
                    }
                }

                markdown.isTrusted = true;
                return new vscode.Hover(markdown);
            }
        }

        this.outputChannel.appendLine('No hover info found, returning null');
        return null;
    }

    private findCustomType(document: vscode.TextDocument, typeName: string):
        { kind: string; line: number; doc?: string; fields?: string[]; values?: string[]; options?: string[] } | null {

        const text = document.getText();
        const lines = text.split('\n');

        // Look for type, enum, or union definitions
        let docComment = '';

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();

            // Collect documentation comments
            if (line.startsWith('///')) {
                docComment += line.substring(3).trim() + ' ';
                continue;
            }

            // Skip leading annotations (they're on their own lines)
            if (line.startsWith('@')) {
                continue;
            }

            // Check for type definition (supports both inline and separate-line annotations)
            const typeMatch = line.match(/^type\s+(\w+)(?:\s+@\w+(?:\.\w+)?(?:\([^)]*\))?)?\s*{/);
            if (typeMatch && typeMatch[1] === typeName) {
                // Extract fields
                const fields: string[] = [];
                let j = i + 1;
                while (j < lines.length) {
                    const fieldLine = lines[j].trim();
                    if (fieldLine === '}') break;

                    const fieldMatch = fieldLine.match(/^(\w+):/);
                    if (fieldMatch) {
                        fields.push(fieldMatch[1]);
                    }
                    j++;
                }

                return {
                    kind: 'type',
                    line: i,
                    doc: docComment.trim() || undefined,
                    fields: fields.length > 0 ? fields : undefined
                };
            }

            // Check for enum definition
            const enumMatch = line.match(/^enum\s+(\w+)\s*{/);
            if (enumMatch && enumMatch[1] === typeName) {
                // Extract enum values
                const values: string[] = [];
                let j = i + 1;
                while (j < lines.length) {
                    const valueLine = lines[j].trim();
                    if (valueLine === '}') break;

                    const valueMatch = valueLine.match(/^(\w+)/);
                    if (valueMatch) {
                        values.push(valueMatch[1]);
                    }
                    j++;
                }

                return {
                    kind: 'enum',
                    line: i,
                    doc: docComment.trim() || undefined,
                    values: values.length > 0 ? values : undefined
                };
            }

            // Check for union definition
            const unionMatch = line.match(/^union\s+(\w+)\s*{/);
            if (unionMatch && unionMatch[1] === typeName) {
                // Extract union options
                const options: string[] = [];
                let j = i + 1;
                while (j < lines.length) {
                    const optionLine = lines[j].trim();
                    if (optionLine === '}') break;

                    const optionMatch = optionLine.match(/^(\w+)/);
                    if (optionMatch) {
                        options.push(optionMatch[1]);
                    }
                    j++;
                }

                return {
                    kind: 'union',
                    line: i,
                    doc: docComment.trim() || undefined,
                    options: options.length > 0 ? options : undefined
                };
            }

            // Reset doc comment if we're not on a comment line
            if (!line.startsWith('///') && line !== '') {
                docComment = '';
            }
        }

        return null;
    }
}
