import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface FieldDefinition {
    name: string;
    type: string;
    required: boolean;
    fieldNumber?: number;
}

export async function newTypeWizard() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'typemux-schema') {
        vscode.window.showErrorMessage('Please open a .typemux file first');
        return;
    }

    // Step 1: Ask for type name
    const typeName = await vscode.window.showInputBox({
        prompt: 'Enter the type name',
        placeHolder: 'User',
        validateInput: (value) => {
            if (!value) return 'Type name is required';
            if (!/^[A-Z][a-zA-Z0-9]*$/.test(value)) {
                return 'Type name must start with a capital letter and contain only letters and numbers';
            }
            return undefined;
        }
    });

    if (!typeName) return;

    // Step 2: Ask if they want to add a documentation comment
    const addDoc = await vscode.window.showQuickPick(['Yes', 'No'], {
        placeHolder: 'Add documentation comment?'
    });

    let docComment = '';
    if (addDoc === 'Yes') {
        const doc = await vscode.window.showInputBox({
            prompt: 'Enter documentation comment',
            placeHolder: 'Description of this type'
        });
        if (doc) {
            docComment = `/// ${doc}\n`;
        }
    }

    // Step 3: Collect fields
    const fields: FieldDefinition[] = [];
    let addingFields = true;
    let fieldNumber = 1;

    while (addingFields) {
        const addField = await vscode.window.showQuickPick(
            fields.length === 0
                ? ['Add field', 'Finish (no fields)']
                : ['Add another field', 'Finish'],
            {
                placeHolder: fields.length === 0
                    ? 'Add fields to the type?'
                    : `Added ${fields.length} field(s). Add more?`
            }
        );

        if (!addField || addField.startsWith('Finish')) {
            addingFields = false;
            break;
        }

        // Get field name
        const fieldName = await vscode.window.showInputBox({
            prompt: `Field ${fields.length + 1}: Enter field name`,
            placeHolder: 'fieldName',
            validateInput: (value) => {
                if (!value) return 'Field name is required';
                if (!/^[a-z][a-zA-Z0-9]*$/.test(value)) {
                    return 'Field name must start with a lowercase letter';
                }
                if (fields.some(f => f.name === value)) {
                    return 'Field name already exists';
                }
                return undefined;
            }
        });

        if (!fieldName) {
            addingFields = false;
            break;
        }

        // Get field type
        const fieldType = await vscode.window.showQuickPick([
            'string',
            'int32',
            'int64',
            'float32',
            'float64',
            'bool',
            'bytes',
            'timestamp',
            '[]string (array)',
            'map<string, string>',
            'Custom type...'
        ], {
            placeHolder: `Field ${fields.length + 1}: Select type for "${fieldName}"`
        });

        if (!fieldType) {
            addingFields = false;
            break;
        }

        let actualType = fieldType;
        if (fieldType === 'Custom type...') {
            const customType = await vscode.window.showInputBox({
                prompt: 'Enter custom type name',
                placeHolder: 'User',
                validateInput: (value) => {
                    if (!value) return 'Type name is required';
                    return undefined;
                }
            });
            if (!customType) {
                addingFields = false;
                break;
            }
            actualType = customType;
        } else if (fieldType.includes('(')) {
            // Remove description text like "(array)"
            actualType = fieldType.split(' (')[0];
        }

        // Ask if field is required
        const isRequired = await vscode.window.showQuickPick(['Yes', 'No'], {
            placeHolder: `Is "${fieldName}" required?`
        });

        if (!isRequired) {
            addingFields = false;
            break;
        }

        fields.push({
            name: fieldName,
            type: actualType,
            required: isRequired === 'Yes',
            fieldNumber: fieldNumber++
        });
    }

    // Generate the type definition
    let typeDefinition = docComment;
    typeDefinition += `type ${typeName} {\n`;

    if (fields.length === 0) {
        typeDefinition += '\t// Add fields here\n';
    } else {
        fields.forEach(field => {
            const requiredAnnotation = field.required ? ' @required' : '';
            typeDefinition += `\t${field.name}: ${field.type} = ${field.fieldNumber}${requiredAnnotation}\n`;
        });
    }

    typeDefinition += '}\n';

    // Insert at current cursor position
    const position = editor.selection.active;
    await editor.edit(editBuilder => {
        editBuilder.insert(position, typeDefinition);
    });

    // Show success message
    vscode.window.showInformationMessage(`✅ Created type "${typeName}" with ${fields.length} field(s)`);
}

export async function newServiceWizard() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'typemux-schema') {
        vscode.window.showErrorMessage('Please open a .typemux file first');
        return;
    }

    // Step 1: Ask for service name
    const serviceName = await vscode.window.showInputBox({
        prompt: 'Enter the service name',
        placeHolder: 'UserService',
        validateInput: (value) => {
            if (!value) return 'Service name is required';
            if (!/^[A-Z][a-zA-Z0-9]*$/.test(value)) {
                return 'Service name must start with a capital letter';
            }
            return undefined;
        }
    });

    if (!serviceName) return;

    // Step 2: Ask for service type
    const serviceType = await vscode.window.showQuickPick([
        'REST API (CRUD)',
        'GraphQL (Query/Mutation/Subscription)',
        'gRPC Service',
        'Custom (empty service)'
    ], {
        placeHolder: 'What type of service?'
    });

    if (!serviceType) return;

    // Step 3: If REST or GraphQL, ask for entity name
    let entityName = '';
    if (serviceType.startsWith('REST') || serviceType.startsWith('GraphQL')) {
        const entityInput = await vscode.window.showInputBox({
            prompt: 'Enter the entity name (singular)',
            placeHolder: 'User',
            validateInput: (value) => {
                if (!value) return 'Entity name is required';
                if (!/^[A-Z][a-zA-Z0-9]*$/.test(value)) {
                    return 'Entity name must start with a capital letter';
                }
                return undefined;
            }
        });

        if (!entityInput) return;
        entityName = entityInput;
    }

    // Generate service definition
    let serviceDefinition = `/// ${serviceName} - Service for managing ${entityName || 'resources'}\n`;
    serviceDefinition += `service ${serviceName} {\n`;

    if (serviceType.startsWith('REST')) {
        const entityLower = entityName.toLowerCase();
        const entityPlural = entityLower + 's';

        serviceDefinition += `\t/// Create a new ${entityLower}\n`;
        serviceDefinition += `\trpc Create(Create${entityName}Request) returns (Create${entityName}Response)\n`;
        serviceDefinition += `\t\t@http(POST)\n`;
        serviceDefinition += `\t\t@path("/api/v1/${entityPlural}")\n`;
        serviceDefinition += `\t\t@graphql(mutation)\n`;
        serviceDefinition += `\t\t@success(201)\n`;
        serviceDefinition += `\t\t@errors(400,409,500)\n\n`;

        serviceDefinition += `\t/// Get ${entityLower} by ID\n`;
        serviceDefinition += `\trpc Get(Get${entityName}Request) returns (Get${entityName}Response)\n`;
        serviceDefinition += `\t\t@http(GET)\n`;
        serviceDefinition += `\t\t@path("/api/v1/${entityPlural}/{id}")\n`;
        serviceDefinition += `\t\t@graphql(query)\n`;
        serviceDefinition += `\t\t@errors(404,500)\n\n`;

        serviceDefinition += `\t/// Update ${entityLower}\n`;
        serviceDefinition += `\trpc Update(Update${entityName}Request) returns (Update${entityName}Response)\n`;
        serviceDefinition += `\t\t@http(PUT)\n`;
        serviceDefinition += `\t\t@path("/api/v1/${entityPlural}/{id}")\n`;
        serviceDefinition += `\t\t@graphql(mutation)\n`;
        serviceDefinition += `\t\t@errors(400,404,500)\n\n`;

        serviceDefinition += `\t/// Delete ${entityLower}\n`;
        serviceDefinition += `\trpc Delete(Delete${entityName}Request) returns (Delete${entityName}Response)\n`;
        serviceDefinition += `\t\t@http(DELETE)\n`;
        serviceDefinition += `\t\t@path("/api/v1/${entityPlural}/{id}")\n`;
        serviceDefinition += `\t\t@graphql(mutation)\n`;
        serviceDefinition += `\t\t@success(204)\n`;
        serviceDefinition += `\t\t@errors(404,500)\n`;

    } else if (serviceType.startsWith('GraphQL')) {
        const entityLower = entityName.toLowerCase();

        serviceDefinition += `\t/// Query: Get ${entityLower} by ID\n`;
        serviceDefinition += `\trpc Get${entityName}(Get${entityName}Request) returns (Get${entityName}Response)\n`;
        serviceDefinition += `\t\t@graphql(query)\n\n`;

        serviceDefinition += `\t/// Query: List all ${entityLower}s\n`;
        serviceDefinition += `\trpc List${entityName}s(List${entityName}sRequest) returns (List${entityName}sResponse)\n`;
        serviceDefinition += `\t\t@graphql(query)\n\n`;

        serviceDefinition += `\t/// Mutation: Create ${entityLower}\n`;
        serviceDefinition += `\trpc Create${entityName}(Create${entityName}Request) returns (Create${entityName}Response)\n`;
        serviceDefinition += `\t\t@graphql(mutation)\n\n`;

        serviceDefinition += `\t/// Subscription: Watch ${entityLower} updates\n`;
        serviceDefinition += `\trpc Watch${entityName}(Watch${entityName}Request) returns (Watch${entityName}Response)\n`;
        serviceDefinition += `\t\t@graphql(subscription)\n`;

    } else if (serviceType.startsWith('gRPC')) {
        serviceDefinition += `\trpc Method(Request) returns (Response)\n`;
    } else {
        serviceDefinition += `\t// Add RPC methods here\n`;
    }

    serviceDefinition += '}\n';

    // Insert at current cursor position
    const position = editor.selection.active;
    await editor.edit(editBuilder => {
        editBuilder.insert(position, serviceDefinition);
    });

    vscode.window.showInformationMessage(`✅ Created ${serviceType} service "${serviceName}"`);
}

export async function newFileWizard() {
    // Ask for file name
    const fileName = await vscode.window.showInputBox({
        prompt: 'Enter file name (without extension)',
        placeHolder: 'users',
        validateInput: (value) => {
            if (!value) return 'File name is required';
            if (!/^[a-z][a-z0-9_-]*$/.test(value)) {
                return 'File name must start with a lowercase letter and contain only lowercase letters, numbers, hyphens, and underscores';
            }
            return undefined;
        }
    });

    if (!fileName) return;

    // Ask for namespace
    const namespace = await vscode.window.showInputBox({
        prompt: 'Enter namespace',
        placeHolder: 'com.example.api',
        validateInput: (value) => {
            if (!value) return 'Namespace is required';
            if (!/^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*$/.test(value)) {
                return 'Namespace must be in format: com.example.api (lowercase, dot-separated)';
            }
            return undefined;
        }
    });

    if (!namespace) return;

    // Ask for description
    const description = await vscode.window.showInputBox({
        prompt: 'Enter a brief description',
        placeHolder: 'User management API'
    });

    // Generate file content
    let content = '@typemux("1.0.0")\n\n';
    content += `/// ${description || 'Schema definition for ' + fileName}\n`;
    content += `namespace ${namespace}\n\n`;
    content += '// Add your types, enums, and services here\n';

    // Create new file
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        vscode.window.showErrorMessage('No workspace folder open');
        return;
    }

    const filePath = vscode.Uri.joinPath(workspaceFolder.uri, `${fileName}.typemux`);

    await vscode.workspace.fs.writeFile(filePath, Buffer.from(content, 'utf8'));

    // Open the new file
    const document = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(document);

    vscode.window.showInformationMessage(`✅ Created new file "${fileName}.typemux"`);
}

export async function importFromExternalFormat() {
    // Step 1: Ask which format to import from
    const format = await vscode.window.showQuickPick([
        {
            label: 'Protocol Buffers (.proto)',
            description: 'Import from Protobuf schema files',
            value: 'proto'
        },
        {
            label: 'GraphQL (.graphql, .gql)',
            description: 'Import from GraphQL schema files',
            value: 'graphql'
        },
        {
            label: 'OpenAPI (.yaml, .yml, .json)',
            description: 'Import from OpenAPI 3.0 specification files',
            value: 'openapi'
        }
    ], {
        placeHolder: 'Select the format to import from'
    });

    if (!format) return;

    // Step 2: Select the input file
    const fileFilters: { [key: string]: string[] } = {
        'proto': ['proto'],
        'graphql': ['graphql', 'gql', 'graphqls'],
        'openapi': ['yaml', 'yml', 'json']
    };

    const inputFiles = await vscode.window.showOpenDialog({
        canSelectFiles: true,
        canSelectFolders: false,
        canSelectMany: false,
        filters: {
            [format.label]: fileFilters[format.value]
        },
        title: `Select ${format.label} file to import`
    });

    if (!inputFiles || inputFiles.length === 0) return;
    const inputFile = inputFiles[0].fsPath;

    // Step 3: Ask for output directory
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        vscode.window.showErrorMessage('No workspace folder open');
        return;
    }

    const outputFolders = await vscode.window.showOpenDialog({
        canSelectFiles: false,
        canSelectFolders: true,
        canSelectMany: false,
        defaultUri: workspaceFolder.uri,
        title: 'Select output directory for generated TypeMux files'
    });

    if (!outputFolders || outputFolders.length === 0) return;
    const outputDir = outputFolders[0].fsPath;

    // Step 4: Find the TypeMux binary path
    // Try to find it relative to the workspace
    const workspacePath = workspaceFolder.uri.fsPath;
    let binaryPath = '';

    // Check if we're in the TypeMux project itself
    const localBinPath = path.join(workspacePath, 'bin', `${format.value}2typemux`);
    if (fs.existsSync(localBinPath)) {
        binaryPath = localBinPath;
    } else {
        // Try to use it from PATH
        binaryPath = `${format.value}2typemux`;
    }

    // Step 5: Run the importer
    const statusMessage = vscode.window.setStatusBarMessage(`$(sync~spin) Importing from ${format.label}...`);

    try {
        const command = `"${binaryPath}" -input "${inputFile}" -output "${outputDir}"`;
        const { stdout, stderr } = await execAsync(command, {
            cwd: workspacePath
        });

        statusMessage.dispose();

        if (stderr && !stderr.includes('warning')) {
            vscode.window.showErrorMessage(`Import failed: ${stderr}`);
            return;
        }

        // Success! Show the output and offer to open files
        const openFiles = await vscode.window.showInformationMessage(
            `✅ Successfully imported from ${format.label}!\n\nFiles generated in: ${outputDir}`,
            'Open Files',
            'Show in Explorer'
        );

        if (openFiles === 'Open Files') {
            // Find and open the generated .typemux files
            const files = await vscode.workspace.findFiles(
                new vscode.RelativePattern(outputDir, '*.typemux'),
                null,
                10
            );

            if (files.length > 0) {
                for (const file of files) {
                    const document = await vscode.workspace.openTextDocument(file);
                    await vscode.window.showTextDocument(document, { preview: false });
                }
            }
        } else if (openFiles === 'Show in Explorer') {
            await vscode.commands.executeCommand('revealInExplorer', vscode.Uri.file(outputDir));
        }

    } catch (error: any) {
        statusMessage.dispose();
        vscode.window.showErrorMessage(
            `Failed to import: ${error.message}\n\nMake sure the TypeMux importers are installed and in your PATH.`
        );
    }
}
