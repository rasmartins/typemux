# frozen_string_literal: true

module Rouge
  module Lexers
    class TypeMux < RegexLexer
      title "TypeMux"
      desc "TypeMux IDL (Interface Definition Language)"
      tag 'typemux'
      filenames '*.typemux'
      mimetypes 'text/x-typemux'

      # Keywords
      keywords = %w(
        type enum union service rpc returns import namespace
      )

      # Built-in types
      builtins = %w(
        string int32 int64 float32 float64 bool timestamp bytes map
      )

      # Annotations
      annotations = %w(
        required default exclude only http path graphql success errors
        proto openapi name option directive extension
      )

      state :root do
        # Comments
        rule %r(///.*?$), Comment::Doc
        rule %r(//.*?$), Comment::Single
        rule %r(/\*), Comment::Multiline, :multiline_comment

        # Keywords
        rule %r/\b(#{keywords.join('|')})\b/, Keyword

        # Built-in types
        rule %r/\b(#{builtins.join('|')})\b/, Keyword::Type

        # Annotations
        rule %r/@(#{annotations.join('|')})\b/, Name::Decorator
        rule %r/@[a-zA-Z_][a-zA-Z0-9_.]*/, Name::Decorator

        # Numbers
        rule %r/\b\d+\b/, Num::Integer

        # Strings
        rule %r/"([^"\\]|\\.)*"/, Str::Double

        # Identifiers (type names, field names)
        rule %r/[A-Z][a-zA-Z0-9_]*/, Name::Class
        rule %r/[a-z_][a-zA-Z0-9_]*/, Name

        # Operators and punctuation
        rule %r/[{}\[\]()<>=:,.]/, Punctuation

        # Whitespace
        rule %r/\s+/, Text
      end

      state :multiline_comment do
        rule %r/\*\//, Comment::Multiline, :pop!
        rule %r/[^*]+/, Comment::Multiline
        rule %r/\*/, Comment::Multiline
      end
    end
  end
end
