/** @type {import("eslint").Linter.Config[]} */
export default [
  {
    files: ["src/**/*.js"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
    },
  },
];
