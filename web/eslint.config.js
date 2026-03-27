import js from '@eslint/js'
import ts from 'typescript-eslint'
import svelte from 'eslint-plugin-svelte'
import svelteParser from 'svelte-eslint-parser'

export default [
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs['flat/recommended'],
  {
    files: ['**/*.svelte', '**/*.svelte.ts'],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: ts.parser,
      },
      globals: {
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        fetch: 'readonly',
        setTimeout: 'readonly',
        clearTimeout: 'readonly',
        setInterval: 'readonly',
        clearInterval: 'readonly',
        console: 'readonly',
        URL: 'readonly',
        HTMLElement: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLSelectElement: 'readonly',
        HTMLTextAreaElement: 'readonly',
        HTMLVideoElement: 'readonly',
        HTMLImageElement: 'readonly',
        KeyboardEvent: 'readonly',
        MouseEvent: 'readonly',
        Event: 'readonly',
        EventSource: 'readonly',
        MediaSource: 'readonly',
        SourceBuffer: 'readonly',
        WebSocket: 'readonly',
        Image: 'readonly',
        ArrayBuffer: 'readonly',
        DragEvent: 'readonly',
        DataTransfer: 'readonly',
        FormData: 'readonly',
        Response: 'readonly',
        RequestInit: 'readonly',
        globalThis: 'readonly',
      },
    },
  },
  {
    files: ['**/*.ts'],
    languageOptions: {
      globals: {
        window: 'readonly',
        document: 'readonly',
        fetch: 'readonly',
        console: 'readonly',
        globalThis: 'readonly',
        setTimeout: 'readonly',
        clearTimeout: 'readonly',
        Response: 'readonly',
        RequestInit: 'readonly',
      },
    },
  },
  {
    rules: {
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
      // Relax for existing codebase
      'svelte/require-each-key': 'warn',
      'svelte/no-at-html-tags': 'warn',
      'svelte/prefer-svelte-reactivity': 'warn',
      'svelte/prefer-writable-derived': 'warn',
      'svelte/no-unused-svelte-ignore': 'warn',
      'no-undef': 'off', // TypeScript handles this
      'no-empty': 'warn',
      'no-case-declarations': 'warn',
      'prefer-const': 'warn',
      // Code complexity limits
      'max-lines': ['warn', { max: 300, skipBlankLines: true, skipComments: true }],
      'max-lines-per-function': ['warn', { max: 40, skipBlankLines: true, skipComments: true }],
      'max-depth': ['warn', 4],
      'complexity': ['warn', 10],
    },
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts'],
    rules: {
      'prefer-const': 'off', // Svelte 5 $props() destructuring requires let, not const
    },
  },
  {
    files: ['**/*.test.ts', '**/*.test.js', '**/*.spec.ts', '**/*.spec.js'],
    rules: {
      'max-lines-per-function': 'off',
      'max-lines': 'off',
    },
  },
  {
    ignores: ['static/', 'node_modules/', '*.config.js', '*.config.ts'],
  },
]
