import * as vscode from 'vscode';
import { TypeMuxFormattingProvider } from './formatter';
import { TypeMuxDefinitionProvider } from './definitionProvider';
import { TypeMuxHoverProvider } from './hoverProvider';

export function activate(context: vscode.ExtensionContext) {
    const outputChannel = vscode.window.createOutputChannel('TypeMux');
    outputChannel.show(); // Show the output channel
    outputChannel.appendLine('=================================');
    outputChannel.appendLine('TypeMux extension v0.4.0 activated');
    outputChannel.appendLine('Build timestamp: ' + new Date().toISOString());
    outputChannel.appendLine('Features: Dot notation annotations support');
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
}

export function deactivate() {}
