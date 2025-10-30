import * as vscode from 'vscode';
import * as path from 'path';
import { SchemaParser, ParsedSchema } from './schemaParser';

export class VisualEditorPanel {
    public static currentPanel: VisualEditorPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private _disposables: vscode.Disposable[] = [];
    private _document: vscode.TextDocument;
    private _parser: SchemaParser;
    private _updating: boolean = false;

    public static createOrShow(extensionUri: vscode.Uri, document: vscode.TextDocument) {
        const column = vscode.ViewColumn.Beside;

        // If we already have a panel, show it
        if (VisualEditorPanel.currentPanel) {
            VisualEditorPanel.currentPanel._panel.reveal(column);
            VisualEditorPanel.currentPanel._document = document;
            VisualEditorPanel.currentPanel.update();
            return;
        }

        // Otherwise, create a new panel
        const panel = vscode.window.createWebviewPanel(
            'typemuxVisualEditor',
            'TypeMux Visual Editor',
            column,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
                localResourceRoots: [extensionUri]
            }
        );

        VisualEditorPanel.currentPanel = new VisualEditorPanel(panel, extensionUri, document);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, document: vscode.TextDocument) {
        this._panel = panel;
        this._extensionUri = extensionUri;
        this._document = document;
        this._parser = new SchemaParser();

        // Set the webview's initial html content
        this._update();

        // Listen for when the panel is disposed
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        // Update the content based on view state changes
        this._panel.onDidChangeViewState(
            e => {
                if (this._panel.visible) {
                    this._update();
                }
            },
            null,
            this._disposables
        );

        // Handle messages from the webview
        this._panel.webview.onDidReceiveMessage(
            message => {
                switch (message.type) {
                    case 'updateText':
                        this.updateTextDocument(message.schema);
                        return;
                    case 'addType':
                        this.addType(message.name);
                        return;
                    case 'deleteType':
                        this.deleteType(message.name);
                        return;
                    case 'addField':
                        this.addField(message.typeName, message.field);
                        return;
                    case 'deleteField':
                        this.deleteField(message.typeName, message.fieldName);
                        return;
                    case 'addImport':
                        this.addImport(message.path);
                        return;
                    case 'deleteImport':
                        this.deleteImport(message.path);
                        return;
                    case 'addTypeAnnotation':
                        this.addTypeAnnotation(message.typeName, message.annotation);
                        return;
                    case 'deleteTypeAnnotation':
                        this.deleteTypeAnnotation(message.typeName, message.annotationName);
                        return;
                    case 'addFieldAnnotation':
                        this.addFieldAnnotation(message.typeName, message.fieldName, message.annotation);
                        return;
                    case 'addMethodAnnotation':
                        this.addMethodAnnotation(message.serviceName, message.methodName, message.annotation);
                        return;
                    case 'deleteMethodAnnotation':
                        this.deleteMethodAnnotation(message.serviceName, message.methodName, message.annotationName);
                        return;
                    case 'pickImportFile':
                        this.pickImportFile();
                        return;
                }
            },
            null,
            this._disposables
        );

        // Listen for document changes
        vscode.workspace.onDidChangeTextDocument(
            e => {
                if (e.document === this._document && !this._updating) {
                    this.update();
                }
            },
            null,
            this._disposables
        );
    }

    public update() {
        this._update();
    }

    private _update() {
        const webview = this._panel.webview;
        this._panel.title = `Visual Editor: ${this._document.fileName.split('/').pop()}`;
        this._panel.webview.html = this._getHtmlForWebview(webview);
    }

    private async updateTextDocument(schema: ParsedSchema) {
        if (this._updating) return;
        this._updating = true;

        try {
            const newText = this.schemaToText(schema);
            const edit = new vscode.WorkspaceEdit();
            const fullRange = new vscode.Range(
                this._document.positionAt(0),
                this._document.positionAt(this._document.getText().length)
            );
            edit.replace(this._document.uri, fullRange, newText);
            await vscode.workspace.applyEdit(edit);
        } finally {
            setTimeout(() => {
                this._updating = false;
            }, 100);
        }
    }

    private schemaToText(schema: ParsedSchema): string {
        let text = `@typemux("${schema.version}")\n\n`;

        if (schema.namespace) {
            text += `namespace ${schema.namespace}\n\n`;
        }

        if (schema.imports.length > 0) {
            schema.imports.forEach(imp => {
                text += `import "${imp}"\n`;
            });
            text += '\n';
        }

        // Generate types
        schema.types.forEach(type => {
            if (type.documentation) {
                text += `/// ${type.documentation}\n`;
            }
            // Output type-level annotations
            type.annotations.forEach(ann => {
                text += `${ann.name}`;
                if (ann.value) {
                    text += `(${ann.value})`;
                }
                text += '\n';
            });
            text += `type ${type.name} {\n`;
            type.fields.forEach(field => {
                if (field.documentation) {
                    text += `\t/// ${field.documentation}\n`;
                }
                text += `\t${field.name}: ${field.type}`;
                if (field.fieldNumber !== undefined) {
                    text += ` = ${field.fieldNumber}`;
                }
                // Output all field annotations
                field.annotations.forEach(ann => {
                    text += ` ${ann.name}`;
                    if (ann.value) {
                        text += `(${ann.value})`;
                    }
                });
                // Handle @required and @default specially if not in annotations already
                if (field.required && !field.annotations.some(a => a.name === '@required')) {
                    text += ' @required';
                }
                if (field.defaultValue && !field.annotations.some(a => a.name.startsWith('@default'))) {
                    text += ` @default(${field.defaultValue})`;
                }
                text += '\n';
            });
            text += '}\n\n';
        });

        // Generate enums
        schema.enums.forEach(enumDef => {
            if (enumDef.documentation) {
                text += `/// ${enumDef.documentation}\n`;
            }
            text += `enum ${enumDef.name} {\n`;
            enumDef.values.forEach(value => {
                text += `\t${value.name} = ${value.number}\n`;
            });
            text += '}\n\n';
        });

        // Generate services
        schema.services.forEach(service => {
            if (service.documentation) {
                text += `/// ${service.documentation}\n`;
            }
            text += `service ${service.name} {\n`;
            service.methods.forEach(method => {
                if (method.documentation) {
                    text += `\t/// ${method.documentation}\n`;
                }
                text += `\trpc ${method.name}(${method.request}) returns (${method.response})\n`;
                method.annotations.forEach(ann => {
                    text += `\t\t${ann.name}`;
                    if (ann.value) {
                        text += `(${ann.value})`;
                    }
                    text += '\n';
                });
            });
            text += '}\n\n';
        });

        return text;
    }

    private async addType(name: string) {
        const schema = this._parser.parse(this._document.getText());
        schema.types.push({
            name,
            documentation: '',
            fields: [],
            annotations: [],
            lineNumber: 0
        });
        await this.updateTextDocument(schema);
    }

    private async deleteType(name: string) {
        const schema = this._parser.parse(this._document.getText());
        schema.types = schema.types.filter(t => t.name !== name);
        await this.updateTextDocument(schema);
    }

    private async addField(typeName: string, field: any) {
        const schema = this._parser.parse(this._document.getText());
        const type = schema.types.find(t => t.name === typeName);
        if (type) {
            type.fields.push({
                name: field.name,
                type: field.type,
                fieldNumber: field.fieldNumber,
                required: field.required || false,
                documentation: '',
                annotations: []
            });
            await this.updateTextDocument(schema);
        }
    }

    private async deleteField(typeName: string, fieldName: string) {
        const schema = this._parser.parse(this._document.getText());
        const type = schema.types.find(t => t.name === typeName);
        if (type) {
            type.fields = type.fields.filter(f => f.name !== fieldName);
            await this.updateTextDocument(schema);
        }
    }

    private async pickImportFile() {
        const fileUris = await vscode.window.showOpenDialog({
            canSelectFiles: true,
            canSelectFolders: false,
            canSelectMany: false,
            filters: {
                'TypeMux Schema': ['typemux']
            },
            openLabel: 'Import'
        });

        if (fileUris && fileUris.length > 0) {
            const selectedFilePath = fileUris[0].fsPath;
            const currentFilePath = this._document.uri.fsPath;

            // Relativize the path
            const relativePath = this.relativizePath(currentFilePath, selectedFilePath);

            // Add the import
            await this.addImport(relativePath);
        }
    }

    private relativizePath(fromPath: string, toPath: string): string {
        const fromDir = path.dirname(fromPath);
        const relativePath = path.relative(fromDir, toPath);

        // Convert Windows paths to Unix-style for consistency
        return relativePath.split(path.sep).join('/');
    }

    private async addImport(importPath: string) {
        const schema = this._parser.parse(this._document.getText());
        if (!schema.imports.includes(importPath)) {
            schema.imports.push(importPath);
            await this.updateTextDocument(schema);
        }
    }

    private async deleteImport(path: string) {
        const schema = this._parser.parse(this._document.getText());
        schema.imports = schema.imports.filter(imp => imp !== path);
        await this.updateTextDocument(schema);
    }

    private async addTypeAnnotation(typeName: string, annotation: { name: string, value?: string }) {
        const schema = this._parser.parse(this._document.getText());
        const type = schema.types.find(t => t.name === typeName);
        if (type) {
            type.annotations.push(annotation);
            await this.updateTextDocument(schema);
        }
    }

    private async deleteTypeAnnotation(typeName: string, annotationName: string) {
        const schema = this._parser.parse(this._document.getText());
        const type = schema.types.find(t => t.name === typeName);
        if (type) {
            type.annotations = type.annotations.filter(a => a.name !== annotationName);
            await this.updateTextDocument(schema);
        }
    }

    private async addFieldAnnotation(typeName: string, fieldName: string, annotation: { name: string, value?: string }) {
        const schema = this._parser.parse(this._document.getText());
        const type = schema.types.find(t => t.name === typeName);
        if (type) {
            const field = type.fields.find(f => f.name === fieldName);
            if (field) {
                field.annotations.push(annotation);
                await this.updateTextDocument(schema);
            }
        }
    }

    private async addMethodAnnotation(serviceName: string, methodName: string, annotation: { name: string, value?: string }) {
        const schema = this._parser.parse(this._document.getText());
        const service = schema.services.find(s => s.name === serviceName);
        if (service) {
            const method = service.methods.find(m => m.name === methodName);
            if (method) {
                method.annotations.push(annotation);
                await this.updateTextDocument(schema);
            }
        }
    }

    private async deleteMethodAnnotation(serviceName: string, methodName: string, annotationName: string) {
        const schema = this._parser.parse(this._document.getText());
        const service = schema.services.find(s => s.name === serviceName);
        if (service) {
            const method = service.methods.find(m => m.name === methodName);
            if (method) {
                method.annotations = method.annotations.filter(a => a.name !== annotationName);
                await this.updateTextDocument(schema);
            }
        }
    }

    private _getHtmlForWebview(webview: vscode.Webview) {
        const schema = this._parser.parse(this._document.getText());

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TypeMux Visual Editor</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
            padding: 20px;
            margin: 0;
        }

        .header {
            border-bottom: 1px solid var(--vscode-panel-border);
            padding-bottom: 10px;
            margin-bottom: 20px;
        }

        .header h1 {
            margin: 0 0 10px 0;
            font-size: 24px;
        }

        .namespace-info {
            color: var(--vscode-descriptionForeground);
            font-size: 14px;
        }

        .section {
            margin-bottom: 30px;
        }

        .section-title {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }

        .section-title h2 {
            margin: 0;
            font-size: 18px;
        }

        .add-button {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 6px 12px;
            cursor: pointer;
            border-radius: 3px;
            font-size: 13px;
        }

        .add-button:hover {
            background-color: var(--vscode-button-hoverBackground);
        }

        .card {
            background-color: var(--vscode-input-background);
            border: 1px solid var(--vscode-panel-border);
            border-radius: 5px;
            padding: 15px;
            margin-bottom: 15px;
        }

        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }

        .card-title {
            font-size: 16px;
            font-weight: bold;
            color: var(--vscode-symbolIcon-classForeground);
        }

        .card-actions button {
            background: transparent;
            border: none;
            color: var(--vscode-foreground);
            cursor: pointer;
            padding: 4px 8px;
            margin-left: 5px;
            border-radius: 3px;
        }

        .card-actions button:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .field-list {
            margin-top: 10px;
        }

        .field-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px;
            border-left: 3px solid var(--vscode-symbolIcon-fieldForeground);
            background-color: var(--vscode-editor-background);
            margin-bottom: 5px;
            border-radius: 3px;
        }

        .field-info {
            flex: 1;
        }

        .field-name {
            font-weight: 600;
            color: var(--vscode-symbolIcon-fieldForeground);
        }

        .field-type {
            color: var(--vscode-symbolIcon-keywordForeground);
            margin-left: 8px;
        }

        .field-meta {
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
            margin-left: 8px;
        }

        .enum-value {
            padding: 5px 10px;
            background-color: var(--vscode-editor-background);
            border-left: 3px solid var(--vscode-symbolIcon-enumForeground);
            margin-bottom: 5px;
            border-radius: 3px;
        }

        .empty-state {
            text-align: center;
            padding: 40px;
            color: var(--vscode-descriptionForeground);
        }

        input, select {
            background-color: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            padding: 6px 8px;
            border-radius: 3px;
            font-family: inherit;
            font-size: inherit;
        }

        input:focus, select:focus {
            outline: 1px solid var(--vscode-focusBorder);
        }

        .inline-form {
            display: flex;
            gap: 10px;
            margin-top: 10px;
            padding: 10px;
            background-color: var(--vscode-editor-background);
            border-radius: 3px;
        }

        .inline-form input {
            flex: 1;
        }

        .inline-form button {
            padding: 6px 12px;
        }

        .annotations {
            margin: 8px 0;
            padding: 8px;
            background-color: var(--vscode-editor-background);
            border-radius: 3px;
        }

        .annotation-tag {
            display: inline-block;
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 8px;
            margin: 2px;
            border-radius: 3px;
            font-size: 11px;
            font-family: monospace;
        }

        .annotation-tag button {
            background: none;
            border: none;
            color: inherit;
            cursor: pointer;
            padding: 0 0 0 4px;
            font-size: 10px;
        }

        .add-annotation-btn {
            background: transparent;
            border: 1px dashed var(--vscode-panel-border);
            color: var(--vscode-descriptionForeground);
            padding: 2px 8px;
            margin: 2px;
            border-radius: 3px;
            font-size: 11px;
            cursor: pointer;
        }

        .add-annotation-btn:hover {
            border-color: var(--vscode-focusBorder);
            color: var(--vscode-foreground);
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 Visual Schema Editor</h1>
        <div class="namespace-info">
            <strong>Namespace:</strong> ${schema.namespace || '(none)'}
            <br><strong>Version:</strong> ${schema.version}
        </div>
    </div>

    <!-- Imports Section -->
    <div class="section">
        <div class="section-title">
            <h2>Imports</h2>
            <button class="add-button" onclick="showAddImport()">+ Add Import</button>
        </div>
        ${schema.imports.length === 0 ? '<div class="empty-state">No imports defined.</div>' : ''}
        ${schema.imports.map(imp => `
            <div class="field-item">
                <div class="field-info">
                    <span class="field-name">import</span>
                    <span class="field-type">"${imp}"</span>
                </div>
                <button onclick="deleteImport('${imp}')" title="Delete Import">🗑️</button>
            </div>
        `).join('')}
    </div>

    <!-- Types Section -->
    <div class="section">
        <div class="section-title">
            <h2>Types</h2>
            <button class="add-button" onclick="showAddType()">+ Add Type</button>
        </div>
        <div id="addTypeForm" style="display: none;" class="inline-form">
            <input type="text" id="newTypeName" placeholder="TypeName" />
            <button class="add-button" onclick="addType()">Create</button>
            <button class="add-button" onclick="cancelAddType()">Cancel</button>
        </div>
        ${schema.types.length === 0 ? '<div class="empty-state">No types defined. Click "Add Type" to create one.</div>' : ''}
        ${schema.types.map(type => `
            <div class="card" data-type="${type.name}">
                <div class="card-header">
                    <div class="card-title">${type.name}</div>
                    <div class="card-actions">
                        <button onclick="addFieldToType('${type.name}')" title="Add Field">+ Field</button>
                        <button onclick="deleteType('${type.name}')" title="Delete Type">🗑️</button>
                    </div>
                </div>
                ${type.documentation ? `<div style="color: var(--vscode-descriptionForeground); margin-bottom: 10px;">${type.documentation}</div>` : ''}
                ${type.annotations.length > 0 ? `
                    <div class="annotations">
                        <strong style="font-size: 11px; color: var(--vscode-descriptionForeground);">Annotations:</strong>
                        ${type.annotations.map(ann => `
                            <span class="annotation-tag">
                                ${ann.name}${ann.value ? `(${ann.value})` : ''}
                                <button onclick="deleteTypeAnnotation('${type.name}', '${ann.name}')" title="Remove">×</button>
                            </span>
                        `).join('')}
                        <button class="add-annotation-btn" onclick="showAddTypeAnnotation('${type.name}')">+ annotation</button>
                    </div>
                ` : `
                    <div class="annotations">
                        <button class="add-annotation-btn" onclick="showAddTypeAnnotation('${type.name}')">+ Add annotation</button>
                    </div>
                `}
                <div id="addTypeAnnotationForm_${type.name}" style="display: none;" class="inline-form">
                    <input type="text" id="typeAnnotationName_${type.name}" list="typeAnnotations" placeholder="@annotation" style="width: 220px;" />
                    <datalist id="typeAnnotations">
                        <option value="@proto.name">@proto.name(name) - Custom Protobuf name</option>
                        <option value="@graphql.name">@graphql.name(name) - Custom GraphQL name</option>
                        <option value="@openapi.name">@openapi.name(name) - Custom OpenAPI name</option>
                        <option value="@graphql.directive">@graphql.directive(@key(...)) - GraphQL directive</option>
                        <option value="@openapi.extension">@openapi.extension({...}) - OpenAPI extension</option>
                    </datalist>
                    <input type="text" id="typeAnnotationValue_${type.name}" placeholder='value: "Name" or {key:"val"}' style="width: 220px;" />
                    <button class="add-button" onclick="confirmAddTypeAnnotation('${type.name}')">Add</button>
                    <button class="add-button" onclick="cancelAddTypeAnnotation('${type.name}')">Cancel</button>
                </div>
                <div class="field-list">
                    ${type.fields.length === 0 ? '<div style="color: var(--vscode-descriptionForeground); padding: 10px;">No fields</div>' : ''}
                    ${type.fields.map(field => `
                        <div>
                            <div class="field-item">
                                <div class="field-info">
                                    <span class="field-name">${field.name}</span>
                                    <span class="field-type">${field.type}</span>
                                    ${field.fieldNumber !== undefined ? `<span class="field-meta">= ${field.fieldNumber}</span>` : ''}
                                    ${field.required ? '<span class="field-meta">@required</span>' : ''}
                                    ${field.defaultValue ? `<span class="field-meta">@default(${field.defaultValue})</span>` : ''}
                                    ${field.annotations.filter(a => a.name !== '@required' && !a.name.startsWith('@default')).map(ann => `
                                        <span class="field-meta">${ann.name}${ann.value ? `(${ann.value})` : ''}</span>
                                    `).join('')}
                                </div>
                                <div style="display: flex; gap: 5px;">
                                    <button onclick="showAddFieldAnnotation('${type.name}', '${field.name}')" title="Add Annotation" style="font-size: 11px;">@+</button>
                                    <button onclick="deleteField('${type.name}', '${field.name}')" title="Delete Field">🗑️</button>
                                </div>
                            </div>
                            <div id="addFieldAnnotationForm_${type.name}_${field.name}" style="display: none; margin-left: 20px; margin-top: 5px; margin-bottom: 5px;" class="inline-form">
                                <input type="text" id="fieldAnnotationName_${type.name}_${field.name}" list="fieldAnnotations" placeholder="@annotation" style="width: 180px;" />
                                <datalist id="fieldAnnotations">
                                    <option value="@required">@required - Field is required</option>
                                    <option value="@default">@default(value) - Default value</option>
                                    <option value="@proto.option">@proto.option([...]) - Protobuf field option</option>
                                    <option value="@graphql.directive">@graphql.directive(@...) - GraphQL directive</option>
                                    <option value="@openapi.extension">@openapi.extension({...}) - OpenAPI extension</option>
                                    <option value="@deprecated">@deprecated - Mark as deprecated</option>
                                </datalist>
                                <input type="text" id="fieldAnnotationValue_${type.name}_${field.name}" placeholder='value (optional)' style="width: 180px;" />
                                <button class="add-button" onclick="confirmAddFieldAnnotation('${type.name}', '${field.name}')">Add</button>
                                <button class="add-button" onclick="cancelAddFieldAnnotation('${type.name}', '${field.name}')">Cancel</button>
                            </div>
                        </div>
                    `).join('')}
                </div>
                <div id="addFieldForm_${type.name}" style="display: none;" class="inline-form">
                    <input type="text" id="fieldName_${type.name}" placeholder="fieldName" />
                    <input type="text" id="fieldType_${type.name}" list="typeList_${type.name}" placeholder="Type" style="width: 150px;" />
                    <datalist id="typeList_${type.name}">
                        <option value="string">string</option>
                        <option value="int32">int32</option>
                        <option value="int64">int64</option>
                        <option value="float32">float32</option>
                        <option value="float64">float64</option>
                        <option value="bool">bool</option>
                        <option value="bytes">bytes</option>
                        <option value="timestamp">timestamp</option>
                        <option value="[]string">[]string (array)</option>
                        <option value="map<string,string>">map&lt;string,string&gt;</option>
                        ${schema.types.map(t => `<option value="${t.name}">${t.name} (custom type)</option>`).join('')}
                        ${schema.enums.map(e => `<option value="${e.name}">${e.name} (enum)</option>`).join('')}
                        ${schema.unions.map(u => `<option value="${u.name}">${u.name} (union)</option>`).join('')}
                    </datalist>
                    <input type="number" id="fieldNumber_${type.name}" placeholder="Field #" style="width: 80px;" />
                    <label><input type="checkbox" id="fieldRequired_${type.name}" /> Required</label>
                    <button class="add-button" onclick="confirmAddField('${type.name}')">Add</button>
                    <button class="add-button" onclick="cancelAddField('${type.name}')">Cancel</button>
                </div>
            </div>
        `).join('')}
    </div>

    <!-- Enums Section -->
    <div class="section">
        <div class="section-title">
            <h2>Enums</h2>
        </div>
        ${schema.enums.length === 0 ? '<div class="empty-state">No enums defined.</div>' : ''}
        ${schema.enums.map(enumDef => `
            <div class="card">
                <div class="card-header">
                    <div class="card-title">${enumDef.name}</div>
                </div>
                ${enumDef.documentation ? `<div style="color: var(--vscode-descriptionForeground); margin-bottom: 10px;">${enumDef.documentation}</div>` : ''}
                ${enumDef.values.map(value => `
                    <div class="enum-value">
                        <strong>${value.name}</strong> = ${value.number}
                    </div>
                `).join('')}
            </div>
        `).join('')}
    </div>

    <!-- Services Section -->
    <div class="section">
        <div class="section-title">
            <h2>Services</h2>
        </div>
        ${schema.services.length === 0 ? '<div class="empty-state">No services defined.</div>' : ''}
        ${schema.services.map(service => `
            <div class="card">
                <div class="card-header">
                    <div class="card-title">${service.name}</div>
                </div>
                ${service.documentation ? `<div style="color: var(--vscode-descriptionForeground); margin-bottom: 10px;">${service.documentation}</div>` : ''}
                ${service.methods.map(method => `
                    <div>
                        <div class="field-item" style="flex-direction: column; align-items: flex-start;">
                            <div style="display: flex; width: 100%; justify-content: space-between; align-items: center;">
                                <div class="field-info">
                                    <span class="field-name">${method.name}</span>
                                    <span class="field-type">(${method.request}) → ${method.response}</span>
                                </div>
                                <button onclick="showAddMethodAnnotation('${service.name}', '${method.name}')" title="Add Annotation" style="font-size: 11px;">@+</button>
                            </div>
                            ${method.annotations.length > 0 ? `
                                <div style="margin-top: 8px; display: flex; flex-wrap: wrap; gap: 4px;">
                                    ${method.annotations.map(ann => `
                                        <span class="annotation-tag">
                                            ${ann.name}${ann.value ? `(${ann.value})` : ''}
                                            <button onclick="deleteMethodAnnotation('${service.name}', '${method.name}', '${ann.name}')" title="Remove">×</button>
                                        </span>
                                    `).join('')}
                                </div>
                            ` : ''}
                        </div>
                        <div id="addMethodAnnotationForm_${service.name}_${method.name}" style="display: none; margin-left: 20px; margin-top: 5px; margin-bottom: 5px;" class="inline-form">
                            <input type="text" id="methodAnnotationName_${service.name}_${method.name}" list="methodAnnotations" placeholder="@annotation" style="width: 180px;" />
                            <datalist id="methodAnnotations">
                                <option value="@http">@http(GET|POST|PUT|DELETE|PATCH) - HTTP method</option>
                                <option value="@path">@path("/api/v1/...") - URL path</option>
                                <option value="@graphql">@graphql(query|mutation|subscription) - GraphQL operation</option>
                                <option value="@success">@success(201|204) - Success status code</option>
                                <option value="@errors">@errors(400,404,500) - Error status codes</option>
                                <option value="@deprecated">@deprecated - Mark as deprecated</option>
                            </datalist>
                            <input type="text" id="methodAnnotationValue_${service.name}_${method.name}" placeholder='value: GET, "/path", etc.' style="width: 180px;" />
                            <button class="add-button" onclick="confirmAddMethodAnnotation('${service.name}', '${method.name}')">Add</button>
                            <button class="add-button" onclick="cancelAddMethodAnnotation('${service.name}', '${method.name}')">Cancel</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `).join('')}
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        function showAddType() {
            document.getElementById('addTypeForm').style.display = 'flex';
            document.getElementById('newTypeName').focus();
        }

        function cancelAddType() {
            document.getElementById('addTypeForm').style.display = 'none';
            document.getElementById('newTypeName').value = '';
        }

        function addType() {
            const name = document.getElementById('newTypeName').value.trim();
            if (name) {
                vscode.postMessage({ type: 'addType', name });
                cancelAddType();
            }
        }

        function deleteType(name) {
            if (confirm(\`Delete type "\${name}"?\`)) {
                vscode.postMessage({ type: 'deleteType', name });
            }
        }

        function addFieldToType(typeName) {
            document.getElementById('addFieldForm_' + typeName).style.display = 'flex';
        }

        function cancelAddField(typeName) {
            document.getElementById('addFieldForm_' + typeName).style.display = 'none';
            document.getElementById('fieldName_' + typeName).value = '';
            document.getElementById('fieldType_' + typeName).value = '';
            document.getElementById('fieldNumber_' + typeName).value = '';
            document.getElementById('fieldRequired_' + typeName).checked = false;
        }

        function confirmAddField(typeName) {
            const name = document.getElementById('fieldName_' + typeName).value.trim();
            const type = document.getElementById('fieldType_' + typeName).value.trim();
            const fieldNumber = parseInt(document.getElementById('fieldNumber_' + typeName).value) || undefined;
            const required = document.getElementById('fieldRequired_' + typeName).checked;

            if (name && type) {
                vscode.postMessage({
                    type: 'addField',
                    typeName,
                    field: { name, type, fieldNumber, required }
                });
                cancelAddField(typeName);
            }
        }

        function deleteField(typeName, fieldName) {
            if (confirm(\`Delete field "\${fieldName}"?\`)) {
                vscode.postMessage({ type: 'deleteField', typeName, fieldName });
            }
        }

        // Import functions
        function showAddImport() {
            vscode.postMessage({ type: 'pickImportFile' });
        }

        function deleteImport(path) {
            if (confirm(\`Delete import "\${path}"?\`)) {
                vscode.postMessage({ type: 'deleteImport', path });
            }
        }

        // Type annotation functions
        function showAddTypeAnnotation(typeName) {
            document.getElementById('addTypeAnnotationForm_' + typeName).style.display = 'flex';
            document.getElementById('typeAnnotationName_' + typeName).focus();
        }

        function cancelAddTypeAnnotation(typeName) {
            document.getElementById('addTypeAnnotationForm_' + typeName).style.display = 'none';
            document.getElementById('typeAnnotationName_' + typeName).value = '';
            document.getElementById('typeAnnotationValue_' + typeName).value = '';
        }

        function confirmAddTypeAnnotation(typeName) {
            const name = document.getElementById('typeAnnotationName_' + typeName).value.trim();
            const value = document.getElementById('typeAnnotationValue_' + typeName).value.trim();

            if (name) {
                const parsed = parseAnnotation(name, value);
                vscode.postMessage({ type: 'addTypeAnnotation', typeName, annotation: parsed });
                cancelAddTypeAnnotation(typeName);
            }
        }

        function deleteTypeAnnotation(typeName, annotationName) {
            vscode.postMessage({ type: 'deleteTypeAnnotation', typeName, annotationName });
        }

        // Field annotation functions
        function showAddFieldAnnotation(typeName, fieldName) {
            document.getElementById('addFieldAnnotationForm_' + typeName + '_' + fieldName).style.display = 'flex';
            document.getElementById('fieldAnnotationName_' + typeName + '_' + fieldName).focus();
        }

        function cancelAddFieldAnnotation(typeName, fieldName) {
            document.getElementById('addFieldAnnotationForm_' + typeName + '_' + fieldName).style.display = 'none';
            document.getElementById('fieldAnnotationName_' + typeName + '_' + fieldName).value = '';
            document.getElementById('fieldAnnotationValue_' + typeName + '_' + fieldName).value = '';
        }

        function confirmAddFieldAnnotation(typeName, fieldName) {
            const name = document.getElementById('fieldAnnotationName_' + typeName + '_' + fieldName).value.trim();
            const value = document.getElementById('fieldAnnotationValue_' + typeName + '_' + fieldName).value.trim();

            if (name) {
                const parsed = parseAnnotation(name, value);
                vscode.postMessage({ type: 'addFieldAnnotation', typeName, fieldName, annotation: parsed });
                cancelAddFieldAnnotation(typeName, fieldName);
            }
        }

        // Method annotation functions
        function showAddMethodAnnotation(serviceName, methodName) {
            document.getElementById('addMethodAnnotationForm_' + serviceName + '_' + methodName).style.display = 'flex';
            document.getElementById('methodAnnotationName_' + serviceName + '_' + methodName).focus();
        }

        function cancelAddMethodAnnotation(serviceName, methodName) {
            document.getElementById('addMethodAnnotationForm_' + serviceName + '_' + methodName).style.display = 'none';
            document.getElementById('methodAnnotationName_' + serviceName + '_' + methodName).value = '';
            document.getElementById('methodAnnotationValue_' + serviceName + '_' + methodName).value = '';
        }

        function confirmAddMethodAnnotation(serviceName, methodName) {
            const name = document.getElementById('methodAnnotationName_' + serviceName + '_' + methodName).value.trim();
            const value = document.getElementById('methodAnnotationValue_' + serviceName + '_' + methodName).value.trim();

            if (name) {
                const parsed = parseAnnotation(name, value);
                vscode.postMessage({ type: 'addMethodAnnotation', serviceName, methodName, annotation: parsed });
                cancelAddMethodAnnotation(serviceName, methodName);
            }
        }

        function deleteMethodAnnotation(serviceName, methodName, annotationName) {
            vscode.postMessage({ type: 'deleteMethodAnnotation', serviceName, methodName, annotationName });
        }

        // Helper function to parse annotation
        function parseAnnotation(name, value) {
            name = name.trim();
            if (!name.startsWith('@')) {
                name = '@' + name;
            }

            return {
                name: name,
                value: value || undefined
            };
        }
    </script>
</body>
</html>`;
    }

    public dispose() {
        VisualEditorPanel.currentPanel = undefined;

        // Clean up our resources
        this._panel.dispose();

        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) {
                x.dispose();
            }
        }
    }
}
