// PrismJS custom language definition for TypeMux
Prism.languages.typemux = {
  'comment': [
    {
      pattern: /(^|[^\\])\/\/\/.*/,
      lookbehind: true,
      greedy: true,
      alias: 'doc-comment'
    },
    {
      pattern: /(^|[^\\])\/\/.*/,
      lookbehind: true,
      greedy: true
    },
    {
      pattern: /(^|[^\\])\/\*[\s\S]*?\*\//,
      lookbehind: true,
      greedy: true
    }
  ],
  'string': {
    pattern: /"(?:[^"\\]|\\.)*"/,
    greedy: true
  },
  'annotation': {
    pattern: /@[a-zA-Z_][a-zA-Z0-9_.]*(?:\([^)]*\))?/,
    inside: {
      'punctuation': /[@().]/,
      'attr-name': /[a-zA-Z_][a-zA-Z0-9_.]*/
    }
  },
  'keyword': /\b(?:type|enum|union|service|rpc|returns|import|namespace)\b/,
  'builtin': /\b(?:string|int32|int64|float32|float64|bool|timestamp|bytes|map)\b/,
  'class-name': /\b[A-Z][a-zA-Z0-9_]*\b/,
  'number': /\b\d+\b/,
  'operator': /=/,
  'punctuation': /[{}[\]()<>:,.]/
};
