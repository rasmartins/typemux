import * as vscode from 'vscode';

export class TypeMuxCodeActionsProvider implements vscode.CodeActionProvider {
    provideCodeActions(
        document: vscode.TextDocument,
        range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        token: vscode.CancellationToken
    ): vscode.CodeAction[] {
        const actions: vscode.CodeAction[] = [];
        const line = document.lineAt(range.start.line);
        const lineText = line.text;
        const trimmed = lineText.trim();

        // Inside a type definition - offer to add a field
        if (this.isInsideType(document, range.start.line)) {
            const addFieldAction = new vscode.CodeAction(
                '$(add) Add field to type',
                vscode.CodeActionKind.RefactorRewrite
            );
            addFieldAction.command = {
                command: 'typemux.addField',
                title: 'Add Field',
                arguments: [document, range.start.line]
            };
            actions.push(addFieldAction);

            const addRequiredFieldAction = new vscode.CodeAction(
                '$(add) Add required field to type',
                vscode.CodeActionKind.RefactorRewrite
            );
            addRequiredFieldAction.command = {
                command: 'typemux.addRequiredField',
                title: 'Add Required Field',
                arguments: [document, range.start.line]
            };
            actions.push(addRequiredFieldAction);
        }

        // Inside an enum definition - offer to add a value
        if (this.isInsideEnum(document, range.start.line)) {
            const addValueAction = new vscode.CodeAction(
                '$(add) Add enum value',
                vscode.CodeActionKind.RefactorRewrite
            );
            addValueAction.command = {
                command: 'typemux.addEnumValue',
                title: 'Add Enum Value',
                arguments: [document, range.start.line]
            };
            actions.push(addValueAction);
        }

        // Inside a service definition - offer to add an RPC method
        if (this.isInsideService(document, range.start.line)) {
            const addRpcAction = new vscode.CodeAction(
                '$(add) Add RPC method',
                vscode.CodeActionKind.RefactorRewrite
            );
            addRpcAction.command = {
                command: 'typemux.addRpcMethod',
                title: 'Add RPC Method',
                arguments: [document, range.start.line]
            };
            actions.push(addRpcAction);
        }

        // At document level - offer to add types, enums, services
        if (this.isAtDocumentLevel(document, range.start.line)) {
            const addTypeAction = new vscode.CodeAction(
                '$(symbol-class) Add new type',
                vscode.CodeActionKind.RefactorRewrite
            );
            addTypeAction.command = {
                command: 'typemux.addType',
                title: 'Add Type',
                arguments: [document, range.start.line]
            };
            actions.push(addTypeAction);

            const addEnumAction = new vscode.CodeAction(
                '$(symbol-enum) Add new enum',
                vscode.CodeActionKind.RefactorRewrite
            );
            addEnumAction.command = {
                command: 'typemux.addEnum',
                title: 'Add Enum',
                arguments: [document, range.start.line]
            };
            actions.push(addEnumAction);

            const addServiceAction = new vscode.CodeAction(
                '$(symbol-method) Add new service',
                vscode.CodeActionKind.RefactorRewrite
            );
            addServiceAction.command = {
                command: 'typemux.addService',
                title: 'Add Service',
                arguments: [document, range.start.line]
            };
            actions.push(addServiceAction);
        }

        // On a field without @required - offer to add it
        if (this.isFieldLine(lineText) && !lineText.includes('@required')) {
            const makeRequiredAction = new vscode.CodeAction(
                'Make field required',
                vscode.CodeActionKind.QuickFix
            );
            makeRequiredAction.edit = new vscode.WorkspaceEdit();
            const endOfLine = line.range.end;
            makeRequiredAction.edit.insert(document.uri, endOfLine, ' @required');
            actions.push(makeRequiredAction);
        }

        // On a field with type but no field number - offer to add it
        if (this.isFieldLine(lineText) && !this.hasFieldNumber(lineText)) {
            const addFieldNumberAction = new vscode.CodeAction(
                'Add field number',
                vscode.CodeActionKind.QuickFix
            );
            addFieldNumberAction.command = {
                command: 'typemux.addFieldNumber',
                title: 'Add Field Number',
                arguments: [document, range.start.line]
            };
            actions.push(addFieldNumberAction);
        }

        return actions;
    }

    private isInsideType(document: vscode.TextDocument, line: number): boolean {
        return this.isInsideBlock(document, line, /^type\s+\w+\s*\{/);
    }

    private isInsideEnum(document: vscode.TextDocument, line: number): boolean {
        return this.isInsideBlock(document, line, /^enum\s+\w+\s*\{/);
    }

    private isInsideService(document: vscode.TextDocument, line: number): boolean {
        return this.isInsideBlock(document, line, /^service\s+\w+\s*\{/);
    }

    private isInsideBlock(document: vscode.TextDocument, line: number, startPattern: RegExp): boolean {
        let braceCount = 0;
        let foundStart = false;

        for (let i = line; i >= 0; i--) {
            const lineText = document.lineAt(i).text.trim();

            // Count closing braces
            const closeBraces = (lineText.match(/\}/g) || []).length;
            braceCount += closeBraces;

            // Count opening braces
            const openBraces = (lineText.match(/\{/g) || []).length;
            braceCount -= openBraces;

            // Check if we're at the start of the block
            if (startPattern.test(lineText)) {
                foundStart = true;
                break;
            }

            // If we've exited the block, stop
            if (braceCount > 0) {
                return false;
            }
        }

        return foundStart && braceCount <= 0;
    }

    private isAtDocumentLevel(document: vscode.TextDocument, line: number): boolean {
        // Check if we're not inside any block
        return !this.isInsideType(document, line) &&
               !this.isInsideEnum(document, line) &&
               !this.isInsideService(document, line);
    }

    private isFieldLine(lineText: string): boolean {
        const trimmed = lineText.trim();
        return /^\w+\s*:\s*\S+/.test(trimmed);
    }

    private hasFieldNumber(lineText: string): boolean {
        return /=\s*\d+/.test(lineText);
    }
}
