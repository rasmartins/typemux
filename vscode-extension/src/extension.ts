import * as vscode from 'vscode';
import { TypeMuxFormattingProvider } from './formatter';
import { TypeMuxDefinitionProvider } from './definitionProvider';
import { TypeMuxHoverProvider } from './hoverProvider';
import { TypeMuxSchemaTreeProvider } from './schemaTreeProvider';
import { TypeMuxCodeActionsProvider } from './codeActionsProvider';

export function activate(context: vscode.ExtensionContext) {
    const outputChannel = vscode.window.createOutputChannel('TypeMux');
    outputChannel.show(); // Show the output channel
    outputChannel.appendLine('=================================');
    outputChannel.appendLine('TypeMux extension v0.5.0 activated');
    outputChannel.appendLine('Build timestamp: ' + new Date().toISOString());
    outputChannel.appendLine('Features: Namespace support, Schema tree view');
    outputChannel.appendLine('=================================');
    console.log('TypeMux extension activated');

    const selector: vscode.DocumentSelector = {
        language: 'typemux-schema',
        scheme: 'file'
    };

    // Register formatting provider
    const formattingProvider = new TypeMuxFormattingProvider();
    context.subscriptions.push(
        vscode.languages.registerDocumentFormattingEditProvider(
            selector,
            formattingProvider
        )
    );
    outputChannel.appendLine('Formatting provider registered');

    // Register go-to-definition provider for imports and types
    const definitionProvider = new TypeMuxDefinitionProvider(outputChannel);
    context.subscriptions.push(
        vscode.languages.registerDefinitionProvider(
            selector,
            definitionProvider
        )
    );
    outputChannel.appendLine('Definition provider registered');

    // Register hover provider for annotations
    outputChannel.appendLine('Creating hover provider...');
    try {
        const hoverProvider = new TypeMuxHoverProvider(outputChannel);
        outputChannel.appendLine('Hover provider instance created');
        context.subscriptions.push(
            vscode.languages.registerHoverProvider(
                selector,
                hoverProvider
            )
        );
        outputChannel.appendLine('Hover provider registered');
    } catch (error) {
        outputChannel.appendLine(`ERROR creating hover provider: ${error}`);
    }

    // Register schema tree view
    outputChannel.appendLine('Creating schema tree view...');
    const treeProvider = new TypeMuxSchemaTreeProvider(context);
    context.subscriptions.push(
        vscode.window.registerTreeDataProvider('typemuxSchemaTree', treeProvider)
    );
    outputChannel.appendLine('Schema tree view registered');

    // Register command to reveal range in editor
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.revealRange', (range: vscode.Range) => {
            const editor = vscode.window.activeTextEditor;
            if (editor) {
                editor.selection = new vscode.Selection(range.start, range.start);
                editor.revealRange(range, vscode.TextEditorRevealType.InCenter);
            }
        })
    );
    outputChannel.appendLine('Reveal range command registered');

    // Register refresh command for tree view
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.refreshSchemaTree', () => {
            treeProvider.refresh();
        })
    );
    outputChannel.appendLine('Refresh tree command registered');

    // Register code actions provider
    outputChannel.appendLine('Registering code actions provider...');
    const codeActionsProvider = new TypeMuxCodeActionsProvider();
    context.subscriptions.push(
        vscode.languages.registerCodeActionsProvider(
            selector,
            codeActionsProvider,
            {
                providedCodeActionKinds: [
                    vscode.CodeActionKind.QuickFix,
                    vscode.CodeActionKind.RefactorRewrite
                ]
            }
        )
    );
    outputChannel.appendLine('Code actions provider registered');

    // Register code action commands
    registerCodeActionCommands(context);
    outputChannel.appendLine('Code action commands registered');
}

function registerCodeActionCommands(context: vscode.ExtensionContext) {
    // Add field to type
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addField', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = findEndOfBlock(document, line);
            const indent = getIndentation(document, line + 1);
            const fieldSnippet = `${indent}\${1:fieldName}: \${2:string}\n`;

            await editor.insertSnippet(
                new vscode.SnippetString(fieldSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add required field to type
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addRequiredField', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = findEndOfBlock(document, line);
            const indent = getIndentation(document, line + 1);
            const fieldSnippet = `${indent}\${1:fieldName}: \${2:string} @required\n`;

            await editor.insertSnippet(
                new vscode.SnippetString(fieldSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add enum value
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addEnumValue', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = findEndOfBlock(document, line);
            const indent = getIndentation(document, line + 1);
            const nextNumber = getNextEnumNumber(document, line);
            const valueSnippet = `${indent}\${1:VALUE} = ${nextNumber}\n`;

            await editor.insertSnippet(
                new vscode.SnippetString(valueSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add RPC method
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addRpcMethod', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = findEndOfBlock(document, line);
            const indent = getIndentation(document, line + 1);
            const rpcSnippet = `${indent}rpc \${1:MethodName}(\${2:Request}) returns (\${3:Response})\n`;

            await editor.insertSnippet(
                new vscode.SnippetString(rpcSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add new type
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addType', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = line + 1;
            const typeSnippet = '\ntype ${1:Name} {\n\t${2:field}: ${3:string} @required\n}\n';

            await editor.insertSnippet(
                new vscode.SnippetString(typeSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add new enum
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addEnum', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = line + 1;
            const enumSnippet = '\nenum ${1:Name} {\n\t${2:VALUE1} = 1\n\t${3:VALUE2} = 2\n}\n';

            await editor.insertSnippet(
                new vscode.SnippetString(enumSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add new service
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addService', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const insertLine = line + 1;
            const serviceSnippet = '\nservice ${1:ServiceName} {\n\trpc ${2:MethodName}(${3:Request}) returns (${4:Response})\n}\n';

            await editor.insertSnippet(
                new vscode.SnippetString(serviceSnippet),
                new vscode.Position(insertLine, 0)
            );
        })
    );

    // Add field number
    context.subscriptions.push(
        vscode.commands.registerCommand('typemux.addFieldNumber', async (document: vscode.TextDocument, line: number) => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const lineText = document.lineAt(line).text;
            const nextNumber = getNextFieldNumber(document, line);

            // Find position after the type, before any annotations
            const colonIndex = lineText.indexOf(':');
            if (colonIndex === -1) return;

            // Find end of type (before @ or end of line)
            let insertPos = lineText.length;
            const atIndex = lineText.indexOf('@');
            if (atIndex !== -1) {
                insertPos = atIndex;
            }

            // Trim whitespace
            while (insertPos > 0 && /\s/.test(lineText[insertPos - 1])) {
                insertPos--;
            }

            await editor.edit(editBuilder => {
                editBuilder.insert(new vscode.Position(line, insertPos), ` = ${nextNumber}`);
            });
        })
    );
}

function findEndOfBlock(document: vscode.TextDocument, startLine: number): number {
    for (let i = startLine + 1; i < document.lineCount; i++) {
        const lineText = document.lineAt(i).text.trim();
        if (lineText === '}') {
            return i;
        }
    }
    return startLine + 1;
}

function getIndentation(document: vscode.TextDocument, line: number): string {
    if (line >= document.lineCount) return '\t';
    const lineText = document.lineAt(line).text;
    const match = lineText.match(/^(\s*)/);
    return match ? match[1] : '\t';
}

function getNextEnumNumber(document: vscode.TextDocument, startLine: number): number {
    let maxNumber = 0;

    for (let i = startLine + 1; i < document.lineCount; i++) {
        const lineText = document.lineAt(i).text;
        if (lineText.trim() === '}') break;

        const numberMatch = lineText.match(/=\s*(\d+)/);
        if (numberMatch) {
            const num = parseInt(numberMatch[1], 10);
            if (num > maxNumber) {
                maxNumber = num;
            }
        }
    }

    return maxNumber + 1;
}

function getNextFieldNumber(document: vscode.TextDocument, currentLine: number): number {
    let maxNumber = 0;

    // Find the containing type block
    let startLine = currentLine;
    while (startLine > 0) {
        const lineText = document.lineAt(startLine).text.trim();
        if (/^type\s+\w+\s*\{/.test(lineText)) {
            break;
        }
        startLine--;
    }

    // Scan all fields in the type
    for (let i = startLine + 1; i < document.lineCount; i++) {
        const lineText = document.lineAt(i).text;
        if (lineText.trim() === '}') break;

        const numberMatch = lineText.match(/=\s*(\d+)/);
        if (numberMatch) {
            const num = parseInt(numberMatch[1], 10);
            if (num > maxNumber) {
                maxNumber = num;
            }
        }
    }

    return maxNumber + 1;
}

export function deactivate() {}
