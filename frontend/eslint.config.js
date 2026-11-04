import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
    },
    rules: {
      // (11-04 — تعديلات ٧ب، قاعدة دازاي) نوافذ المتصفح ممنوعة: askConfirm/askInput من ui/dialogs
      'no-restricted-globals': ['error',
        { name: 'confirm', message: 'ممنوع confirm() — استعمل await askConfirm() من ui/dialogs' },
        { name: 'prompt', message: 'ممنوع prompt() — استعمل await askInput() من ui/dialogs' }],
      'no-restricted-properties': ['error',
        { object: 'window', property: 'confirm', message: 'ممنوع window.confirm — askConfirm' },
        { object: 'window', property: 'prompt', message: 'ممنوع window.prompt — askInput' }],
    },
  },
])
