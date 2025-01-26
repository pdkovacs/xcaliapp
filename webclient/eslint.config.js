import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import pluginReact from "eslint-plugin-react";


/** @type {import('eslint').Linter.Config[]} */
export default [
  { files: ["**/*.{js,mjs,cjs,ts,jsx,tsx}"] },
  { languageOptions: { globals: globals.browser } },
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
  pluginReact.configs.flat.recommended,
  {
    rules: {
      "semi": 2,
      "comma-dangle": ["error", "never"],
      "object-curly-spacing": ["error", "always"],
      "@typescript-eslint/consistent-type-imports": [
        2,
        {
          "fixStyle": "separate-type-imports"
        }
      ],
      "@typescript-eslint/no-restricted-imports": [
        2,
        {
          "paths": [
            {
              "name": "react-redux",
              "importNames": [
                "useSelector",
                "useStore",
                "useDispatch"
              ],
              "message": "Please use pre-typed versions from `src/app/hooks.ts` instead."
            }
          ]
        }
      ],
      "react/react-in-jsx-scope": "off",
      "react/jsx-uses-react": "off"
    }
  }
];
