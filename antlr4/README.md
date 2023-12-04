pip install antlr4-tools
antlr4 -Dlanguage=golang -listener -visitor -o parsers SimpleExprLexer.g4 SimpleExprParser.g4 SimpleExprLexer.tokens
