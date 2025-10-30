import * as vscode from 'vscode';
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
            text += `type ${type.name} {\n`;
            type.fields.forEach(field => {
                if (field.documentation) {
                    text += `\t/// ${field.documentation}\n`;
                }
                text += `\t${field.name}: ${field.type}`;
                if (field.fieldNumber !== undefined) {
                    text += ` = ${field.fieldNumber}`;
                }
                if (field.required) {
                    text += ' @required';
                }
                if (field.defaultValue) {
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
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 Visual Schema Editor</h1>
        <div class="namespace-info">
            <strong>Namespace:</strong> ${schema.namespace || '(none)'}
            ${schema.imports.length > 0 ? `<br><strong>Imports:</strong> ${schema.imports.join(', ')}` : ''}
        </div>
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
                <div class="field-list">
                    ${type.fields.length === 0 ? '<div style="color: var(--vscode-descriptionForeground); padding: 10px;">No fields</div>' : ''}
                    ${type.fields.map(field => `
                        <div class="field-item">
                            <div class="field-info">
                                <span class="field-name">${field.name}</span>
                                <span class="field-type">${field.type}</span>
                                ${field.fieldNumber !== undefined ? `<span class="field-meta">= ${field.fieldNumber}</span>` : ''}
                                ${field.required ? '<span class="field-meta">@required</span>' : ''}
                                ${field.defaultValue ? `<span class="field-meta">@default(${field.defaultValue})</span>` : ''}
                            </div>
                            <button onclick="deleteField('${type.name}', '${field.name}')" title="Delete Field">🗑️</button>
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
                    <div class="field-item">
                        <div class="field-info">
                            <span class="field-name">${method.name}</span>
                            <span class="field-type">(${method.request}) → ${method.response}</span>
                            ${method.annotations.map(ann => `<span class="field-meta">${ann.name}</span>`).join(' ')}
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
