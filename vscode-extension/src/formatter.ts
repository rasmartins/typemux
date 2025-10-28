import * as vscode from 'vscode';

export class TypeMuxFormattingProvider implements vscode.DocumentFormattingEditProvider {
    provideDocumentFormattingEdits(
        document: vscode.TextDocument,
        options: vscode.FormattingOptions,
        token: vscode.CancellationToken
    ): vscode.TextEdit[] {
        const edits: vscode.TextEdit[] = [];
        const lines = document.getText().split('\n');
        let formatted: string[] = [];
        let indentLevel = 0;
        const indentStr = options.insertSpaces ? ' '.repeat(options.tabSize) : '\t';

        for (let i = 0; i < lines.length; i++) {
            let line = lines[i].trim();

            // Skip empty lines and preserve them
            if (line === '') {
                formatted.push('');
                continue;
            }

            // Handle closing braces - decrease indent before the line
            if (line.startsWith('}')) {
                indentLevel = Math.max(0, indentLevel - 1);
            }

            // Apply indentation
            const indentedLine = indentStr.repeat(indentLevel) + line;
            formatted.push(indentedLine);

            // Handle opening braces - increase indent after the line
            if (line.endsWith('{')) {
                indentLevel++;
            }
        }

        // Create a single edit that replaces the entire document
        const fullRange = new vscode.Range(
            document.positionAt(0),
            document.positionAt(document.getText().length)
        );

        edits.push(vscode.TextEdit.replace(fullRange, formatted.join('\n')));
        return edits;
    }
}
