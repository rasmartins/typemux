import * as vscode from 'vscode';
import * as path from 'path';

export class TypeMuxSchemaTreeProvider implements vscode.TreeDataProvider<SchemaItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<SchemaItem | undefined | null | void> = new vscode.EventEmitter<SchemaItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<SchemaItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private context: vscode.ExtensionContext) {
        // Watch for changes to typemux files
        const watcher = vscode.workspace.createFileSystemWatcher('**/*.typemux');
        watcher.onDidChange(() => this.refresh());
        watcher.onDidCreate(() => this.refresh());
        watcher.onDidDelete(() => this.refresh());
        context.subscriptions.push(watcher);

        // Watch for active editor changes
        vscode.window.onDidChangeActiveTextEditor(() => this.refresh(), null, context.subscriptions);
        vscode.workspace.onDidChangeTextDocument(e => {
            if (e.document.languageId === 'typemux-schema') {
                this.refresh();
            }
        }, null, context.subscriptions);
    }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: SchemaItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: SchemaItem): Promise<SchemaItem[]> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || editor.document.languageId !== 'typemux-schema') {
            return [];
        }

        const document = editor.document;
        const text = document.getText();

        if (!element) {
            // Root level - parse the entire schema
            return this.parseSchema(document, text);
        } else {
            // Return children of the element
            return element.children || [];
        }
    }

    private parseSchema(document: vscode.TextDocument, text: string): SchemaItem[] {
        const items: SchemaItem[] = [];
        const lines = text.split('\n');

        let namespace: string | null = null;
        let currentDoc = '';

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const trimmed = line.trim();

            // Collect documentation comments
            if (trimmed.startsWith('///')) {
                currentDoc += trimmed.substring(3).trim() + ' ';
                continue;
            }

            // Parse namespace
            const namespaceMatch = trimmed.match(/^namespace\s+([\w.]+)/);
            if (namespaceMatch) {
                namespace = namespaceMatch[1];
                const item = new SchemaItem(
                    `namespace ${namespace}`,
                    vscode.TreeItemCollapsibleState.None,
                    'namespace',
                    new vscode.Range(i, 0, i, line.length)
                );
                item.iconPath = new vscode.ThemeIcon('symbol-namespace');
                items.push(item);
                currentDoc = '';
                continue;
            }

            // Parse type (with optional trailing annotations)
            const typeNameMatch = trimmed.match(/^type\s+(\w+)/);
            if (typeNameMatch && trimmed.includes('{')) {
                const typeName = typeNameMatch[1];
                const fields = this.parseFields(lines, i);
                const item = new SchemaItem(
                    typeName,
                    vscode.TreeItemCollapsibleState.Collapsed,
                    'type',
                    new vscode.Range(i, 0, i, line.length),
                    currentDoc.trim() || undefined
                );
                item.iconPath = new vscode.ThemeIcon('symbol-class');
                item.children = fields.map(f => {
                    const fieldItem = new SchemaItem(
                        f.name,
                        vscode.TreeItemCollapsibleState.None,
                        'field',
                        f.range,
                        f.description
                    );
                    fieldItem.iconPath = new vscode.ThemeIcon('symbol-field');
                    return fieldItem;
                });
                items.push(item);
                currentDoc = '';
                continue;
            }

            // Parse enum (with optional trailing annotations)
            const enumNameMatch = trimmed.match(/^enum\s+(\w+)/);
            if (enumNameMatch && trimmed.includes('{')) {
                const enumName = enumNameMatch[1];
                const values = this.parseEnumValues(lines, i);
                const item = new SchemaItem(
                    enumName,
                    vscode.TreeItemCollapsibleState.Collapsed,
                    'enum',
                    new vscode.Range(i, 0, i, line.length),
                    currentDoc.trim() || undefined
                );
                item.iconPath = new vscode.ThemeIcon('symbol-enum');
                item.children = values.map(v => {
                    const valueItem = new SchemaItem(
                        v.name,
                        vscode.TreeItemCollapsibleState.None,
                        'enum-value',
                        v.range
                    );
                    valueItem.iconPath = new vscode.ThemeIcon('symbol-enum-member');
                    return valueItem;
                });
                items.push(item);
                currentDoc = '';
                continue;
            }

            // Parse union (with optional trailing annotations)
            const unionNameMatch = trimmed.match(/^union\s+(\w+)/);
            if (unionNameMatch && trimmed.includes('{')) {
                const unionName = unionNameMatch[1];
                const options = this.parseUnionOptions(lines, i);
                const item = new SchemaItem(
                    unionName,
                    vscode.TreeItemCollapsibleState.Collapsed,
                    'union',
                    new vscode.Range(i, 0, i, line.length),
                    currentDoc.trim() || undefined
                );
                item.iconPath = new vscode.ThemeIcon('symbol-interface');
                item.children = options.map(o => {
                    const optionItem = new SchemaItem(
                        o.name,
                        vscode.TreeItemCollapsibleState.None,
                        'union-option',
                        o.range
                    );
                    optionItem.iconPath = new vscode.ThemeIcon('symbol-constant');
                    return optionItem;
                });
                items.push(item);
                currentDoc = '';
                continue;
            }

            // Parse service (with optional trailing annotations)
            const serviceNameMatch = trimmed.match(/^service\s+(\w+)/);
            if (serviceNameMatch && trimmed.includes('{')) {
                const serviceName = serviceNameMatch[1];
                const methods = this.parseServiceMethods(lines, i);
                const item = new SchemaItem(
                    serviceName,
                    vscode.TreeItemCollapsibleState.Collapsed,
                    'service',
                    new vscode.Range(i, 0, i, line.length),
                    currentDoc.trim() || undefined
                );
                item.iconPath = new vscode.ThemeIcon('symbol-method');
                item.children = methods.map(m => {
                    const methodItem = new SchemaItem(
                        m.name,
                        vscode.TreeItemCollapsibleState.None,
                        'rpc',
                        m.range,
                        m.description
                    );
                    methodItem.iconPath = new vscode.ThemeIcon('symbol-function');
                    return methodItem;
                });
                items.push(item);
                currentDoc = '';
                continue;
            }

            // Reset doc comment if line is not empty and not a comment
            if (trimmed !== '' && !trimmed.startsWith('//')) {
                currentDoc = '';
            }
        }

        return items;
    }

    private parseFields(lines: string[], startLine: number): Array<{name: string, range: vscode.Range, description?: string}> {
        const fields: Array<{name: string, range: vscode.Range, description?: string}> = [];
        let i = startLine + 1;
        let fieldDoc = '';

        while (i < lines.length) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed === '}') break;

            // Collect field documentation
            if (trimmed.startsWith('///')) {
                fieldDoc += trimmed.substring(3).trim() + ' ';
                i++;
                continue;
            }

            // Skip leading annotations (field-level annotations on their own line)
            if (trimmed.startsWith('@')) {
                i++;
                continue;
            }

            // Parse field
            const fieldMatch = trimmed.match(/^(\w+)\s*:/);
            if (fieldMatch) {
                const fieldName = fieldMatch[1];
                const typeMatch = trimmed.match(/:\s*(\S+)/);
                const type = typeMatch ? typeMatch[1].replace(/\s*=.*$/, '').replace(/@.*$/, '').trim() : 'unknown';

                fields.push({
                    name: `${fieldName}: ${type}`,
                    range: new vscode.Range(i, 0, i, line.length),
                    description: fieldDoc.trim() || undefined
                });
                fieldDoc = '';
            }

            i++;
        }

        return fields;
    }

    private parseEnumValues(lines: string[], startLine: number): Array<{name: string, range: vscode.Range}> {
        const values: Array<{name: string, range: vscode.Range}> = [];
        let i = startLine + 1;

        while (i < lines.length) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed === '}') break;

            // Skip comments
            if (trimmed.startsWith('//')) {
                i++;
                continue;
            }

            // Parse enum value
            const valueMatch = trimmed.match(/^(\w+)/);
            if (valueMatch) {
                const valueName = valueMatch[1];
                const numberMatch = trimmed.match(/=\s*(\d+)/);
                const displayName = numberMatch ? `${valueName} = ${numberMatch[1]}` : valueName;

                values.push({
                    name: displayName,
                    range: new vscode.Range(i, 0, i, line.length)
                });
            }

            i++;
        }

        return values;
    }

    private parseUnionOptions(lines: string[], startLine: number): Array<{name: string, range: vscode.Range}> {
        const options: Array<{name: string, range: vscode.Range}> = [];
        let i = startLine + 1;

        while (i < lines.length) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed === '}') break;

            // Skip comments
            if (trimmed.startsWith('//')) {
                i++;
                continue;
            }

            // Parse union option
            const optionMatch = trimmed.match(/^([\w.]+)/);
            if (optionMatch) {
                options.push({
                    name: optionMatch[1],
                    range: new vscode.Range(i, 0, i, line.length)
                });
            }

            i++;
        }

        return options;
    }

    private parseServiceMethods(lines: string[], startLine: number): Array<{name: string, range: vscode.Range, description?: string}> {
        const methods: Array<{name: string, range: vscode.Range, description?: string}> = [];
        let i = startLine + 1;
        let methodDoc = '';

        while (i < lines.length) {
            const line = lines[i];
            const trimmed = line.trim();

            if (trimmed === '}') break;

            // Collect method documentation
            if (trimmed.startsWith('///')) {
                methodDoc += trimmed.substring(3).trim() + ' ';
                i++;
                continue;
            }

            // Skip annotations
            if (trimmed.startsWith('@')) {
                i++;
                continue;
            }

            // Parse RPC method
            const rpcMatch = trimmed.match(/^rpc\s+(\w+)\s*\(([^)]+)\)\s*returns\s*\(([^)]+)\)/);
            if (rpcMatch) {
                const methodName = rpcMatch[1];
                const request = rpcMatch[2];
                const response = rpcMatch[3];

                methods.push({
                    name: `${methodName}(${request}) → ${response}`,
                    range: new vscode.Range(i, 0, i, line.length),
                    description: methodDoc.trim() || undefined
                });
                methodDoc = '';
            }

            i++;
        }

        return methods;
    }
}

class SchemaItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly kind: string,
        public readonly range: vscode.Range,
        public readonly description?: string,
        public children?: SchemaItem[]
    ) {
        super(label, collapsibleState);
        this.tooltip = description || label;
        this.command = {
            command: 'typemux.revealRange',
            title: 'Go to Definition',
            arguments: [range]
        };
    }

    contextValue = this.kind;
}
