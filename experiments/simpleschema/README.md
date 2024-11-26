# Simple Schema

`simpleschema` is a [graphql like](https://stackoverflow.com/questions/51563960/how-to-add-default-values-to-input-arguments-in-graphql) schema to define CRDs spec part.
For the status it allows you to annotate possible [source
values](https://github.com/awslabs/kro/blob/dfa4d35bbdf5d0b6eb79bf3d19a9dcb3d5b11ba0/examples/cel-functions/rg.yaml#L13) for the fields.

The kro.run's implementation is available [here](https://github.com/awslabs/kro/tree/main/internal/simpleschema)

The forked copy here allows us to start evolving into a common package that can be used broadly.
