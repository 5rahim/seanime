import type { CodegenConfig } from "@graphql-codegen/cli"

const config: CodegenConfig = {
    overwrite: true,
    schema: "https://graphql.anilist.co",
    documents: "src/lib/anilist/**/*.graphql",
    generates: {
        "src/lib/anilist/gql/": {
            preset: "client",
            plugins: [],
            config: {
                ignoreNoDocuments: true,
                skipTypename: true,
                enumsAsTypes: true,
                constEnums: true,
                nameSuffix: "",
                scalars: {
                    uuid: "string",
                    timestamptz: "any",
                    jsonb: "any",
                },
            },
            presetConfig: {
                fragmentMasking: false,
            },
        },
    },
}

export default config