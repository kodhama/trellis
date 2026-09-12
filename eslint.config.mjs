import js from "@eslint/js";

export default [
  {
    files: [
      "eslint.config.mjs",
      "plugins/trellis/hooks/codex-context.mjs",
      "scripts/plugin-smoke.mjs",
    ],
    ...js.configs.recommended,
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: {
        Buffer: "readonly",
        console: "readonly",
        process: "readonly",
      },
    },
    rules: {
      "no-constant-condition": ["error", { checkLoops: false }],
    },
  },
];
