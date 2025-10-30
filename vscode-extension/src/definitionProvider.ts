import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

export class TypeMuxDefinitionProvider implements vscode.DefinitionProvider {
    private outputChannel: vscode.OutputChannel;

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
        this.outputChannel.appendLine('TypeMuxDefinitionProvider constructor called');
    }

    provideDefinition(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<vscode.Definition | vscode.LocationLink[]> {
        this.outputChannel.appendLine(`=== provideDefinition called at ${position.line}:${position.character} ===`);
        const line = document.lineAt(position.line);
        const lineText = line.text;
        this.outputChannel.appendLine(`Line text: "${lineText}"`);

        // Check if we're on an import line
        const importMatch = lineText.match(/import\s+"([^"]+)"/);
        if (importMatch) {
            const importPath = importMatch[1];

            // Check if cursor is within the import path string
            const startIndex = lineText.indexOf('"') + 1;
            const endIndex = lineText.lastIndexOf('"');
            if (position.character < startIndex || position.character > endIndex) {
                return null;
            }

            // Resolve the import path relative to the current file
            const currentFileDir = path.dirname(document.uri.fsPath);
            const resolvedPath = path.resolve(currentFileDir, importPath);

            // Check if the file exists
            if (!fs.existsSync(resolvedPath)) {
                return null;
            }

            // Return the location of the imported file
            return new vscode.Location(
                vscode.Uri.file(resolvedPath),
                new vscode.Position(0, 0)
            );
        }

        // Check if we're on a type reference (qualified or unqualified)
        this.outputChannel.appendLine('Attempting to extract qualified identifier...');
        const qualifiedIdentifier = this.getQualifiedIdentifierAtPosition(document, position);
        this.outputChannel.appendLine(`Qualified identifier: ${qualifiedIdentifier ? `"${qualifiedIdentifier}"` : 'null'}`);
        if (!qualifiedIdentifier) {
            this.outputChannel.appendLine('No qualified identifier found, returning null');
            return null;
        }

        // Check if this is a qualified type reference (e.g., com.example.users.User)
        if (qualifiedIdentifier.includes('.')) {
            this.outputChannel.appendLine('Detected qualified type reference (contains dot)');
            // Extract namespace and type name
            const lastDotIndex = qualifiedIdentifier.lastIndexOf('.');
            const namespace = qualifiedIdentifier.substring(0, lastDotIndex);
            const typeName = qualifiedIdentifier.substring(lastDotIndex + 1);
            this.outputChannel.appendLine(`  Namespace: "${namespace}"`);
            this.outputChannel.appendLine(`  Type name: "${typeName}"`);

            // Find the file that declares this namespace
            this.outputChannel.appendLine(`Searching for file with namespace "${namespace}"...`);
            const targetFile = this.findFileByNamespace(document, namespace);
            this.outputChannel.appendLine(`  Target file: ${targetFile ? targetFile : 'not found'}`);
            if (targetFile) {
                // Search for the type in that specific file
                this.outputChannel.appendLine(`Searching for type "${typeName}" in ${targetFile}...`);
                const location = this.findTypeDefinitionInFile(targetFile, typeName);
                if (location) {
                    this.outputChannel.appendLine(`  FOUND at line ${location.range.start.line}`);
                    return location;
                } else {
                    this.outputChannel.appendLine(`  NOT FOUND`);
                }
            }
            this.outputChannel.appendLine('Returning null for qualified reference');
            return null;
        }

        // Unqualified type reference - search in current file first
        this.outputChannel.appendLine('Detected unqualified type reference');
        this.outputChannel.appendLine(`Searching for type "${qualifiedIdentifier}" in current file...`);
        const currentFileLocation = this.findTypeDefinition(document, qualifiedIdentifier);
        if (currentFileLocation) {
            this.outputChannel.appendLine(`  FOUND at line ${currentFileLocation.range.start.line}`);
            return currentFileLocation;
        }

        // Search in imported files
        this.outputChannel.appendLine('Not found in current file, searching imported files...');
        const result = this.searchImportedFiles(document, qualifiedIdentifier);
        this.outputChannel.appendLine(`Result from imported files: ${result ? 'FOUND' : 'not found'}`);
        return result;
    }

    private findTypeDefinition(document: vscode.TextDocument, typeName: string): vscode.Location | null {
        const text = document.getText();
        const lines = text.split('\n');

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            // Match: type TypeName {, enum TypeName {, or service TypeName {
            const match = line.match(new RegExp(`^\\s*(type|enum|service)\\s+${typeName}\\s*\\{`));
            if (match) {
                return new vscode.Location(
                    document.uri,
                    new vscode.Position(i, match[0].indexOf(typeName))
                );
            }
        }

        return null;
    }

    private searchImportedFiles(document: vscode.TextDocument, typeName: string): vscode.Location | null {
        const text = document.getText();
        const importMatches = text.matchAll(/import\s+"([^"]+)"/g);

        const currentFileDir = path.dirname(document.uri.fsPath);

        for (const match of importMatches) {
            const importPath = match[1];
            const resolvedPath = path.resolve(currentFileDir, importPath);

            if (!fs.existsSync(resolvedPath)) {
                continue;
            }

            try {
                const importedContent = fs.readFileSync(resolvedPath, 'utf-8');
                const lines = importedContent.split('\n');

                for (let i = 0; i < lines.length; i++) {
                    const line = lines[i];
                    const typeMatch = line.match(new RegExp(`^\\s*(type|enum|service)\\s+${typeName}\\s*\\{`));
                    if (typeMatch) {
                        return new vscode.Location(
                            vscode.Uri.file(resolvedPath),
                            new vscode.Position(i, typeMatch[0].indexOf(typeName))
                        );
                    }
                }
            } catch (err) {
                // Skip files we can't read
                continue;
            }
        }

        return null;
    }

    private getQualifiedIdentifierAtPosition(document: vscode.TextDocument, position: vscode.Position): string | null {
        const line = document.lineAt(position.line);
        const lineText = line.text;

        // Get the character at the cursor position
        let start = position.character;
        let end = position.character;
        this.outputChannel.appendLine(`  Starting position: ${position.character}, char: "${lineText[position.character]}"`);

        // Expand left to find the start of the qualified identifier
        while (start > 0 && /[a-zA-Z0-9.]/.test(lineText[start - 1])) {
            start--;
        }

        // Expand right to find the end of the qualified identifier
        while (end < lineText.length && /[a-zA-Z0-9.]/.test(lineText[end])) {
            end++;
        }

        const identifier = lineText.substring(start, end);
        this.outputChannel.appendLine(`  Extracted substring (${start}-${end}): "${identifier}"`);

        // Check if this is a valid qualified identifier (namespace.Type or just Type)
        // Qualified: com.example.users.User (lowercase parts separated by dots, ending with capitalized Type)
        // Unqualified: User (just a capitalized type name)
        const isValid = /^([a-z][a-zA-Z0-9]*\.)*[A-Z][a-zA-Z0-9]*$/.test(identifier);
        this.outputChannel.appendLine(`  Validation regex test: ${isValid}`);
        if (isValid) {
            return identifier;
        }

        return null;
    }

    private extractNamespace(filePath: string): string | null {
        try {
            const content = fs.readFileSync(filePath, 'utf-8');
            const namespaceMatch = content.match(/namespace\s+([\w.]+)/);
            return namespaceMatch ? namespaceMatch[1] : null;
        } catch {
            return null;
        }
    }

    private findFileByNamespace(document: vscode.TextDocument, namespace: string): string | null {
        const text = document.getText();
        const importMatches = text.matchAll(/import\s+"([^"]+)"/g);
        const currentFileDir = path.dirname(document.uri.fsPath);

        for (const match of importMatches) {
            const importPath = match[1];
            const resolvedPath = path.resolve(currentFileDir, importPath);

            if (fs.existsSync(resolvedPath)) {
                const fileNamespace = this.extractNamespace(resolvedPath);
                if (fileNamespace === namespace) {
                    return resolvedPath;
                }
            }
        }

        return null;
    }

    private findTypeDefinitionInFile(filePath: string, typeName: string): vscode.Location | null {
        try {
            const content = fs.readFileSync(filePath, 'utf-8');
            const lines = content.split('\n');

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];
                // Match: type TypeName {, enum TypeName {, or service TypeName {
                const match = line.match(new RegExp(`^\\s*(type|enum|service)\\s+${typeName}\\s*\\{`));
                if (match) {
                    return new vscode.Location(
                        vscode.Uri.file(filePath),
                        new vscode.Position(i, match[0].indexOf(typeName))
                    );
                }
            }
        } catch {
            return null;
        }

        return null;
    }
}
