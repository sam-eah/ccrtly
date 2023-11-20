## CCRTLY

Run scripts on multiple environments and/or tenants concurrently (or sequentially)

Based on Cobra (CLI), Viper (Config files) & BubbleTea (TUI)

## Commands

Run CLI

```sh
ccrtly
```

#### TODO

- [ ] Run for all available env and tenants

```sh
ccrtly —-all
```

<!-- Run for selected env/tenants. Use * for all.
```sh
ccrtly —tenant=tenant,tenant-2 —env=dev, stg
```
-> just use the tui -->

- [ ] Skip second phase of choosing action to run (still have to select combos)

```sh
ccrtly --script [SCRIPT_NAME]
```

- [ ] Will run immediately for action (plan) for all combos

```sh
ccrtly --script [SCRIPT_NAME] —-all
```

### Select combos (tenant + env)

With the table:

- Press a to select all
- R to select all row
- C to select all column
- Space to toggle
- Enter to confirm selection

### Select script

With the list, press enter to validate.

## Config

`project.yaml` :

```yml
envs:
  - env: ENV_1
    variables:
      ENV_VAR_1_KEY: ENV_VAR_1_VALUE
    tenants:
      TENANT_1:
        VAR_1_KEY: VAR_1_VALUE
        VAR_2_KEY: VAR_2_VALUE
      TENANT_2:
        VAR_1_KEY: VAR_1_VALUE
        VAR_2_KEY: VAR_2_VALUE
  - env: ENV_2
    tenants:
      TENANT_1:
        VAR_1_KEY: VAR_1_VALUE
        VAR_2_KEY: VAR_2_VALUE

prescript: |
  echo "before"

postscript: |
  echo "after"

scripts:
  - name: script1
    command: |
      echo "script1 $CCRTLY_ENV $CCRTLY_TENANT"
      ls
  - name: script2
    command: |
      echo "script2 $CCRTLY_ENV $CCRTLY_TENANT"
      ls

tenants:
  TENANT_1:
    TENANT_VAR_1_KEY: TENANT_VAR_1_VALUE
    TENANT_VAR_2_KEY: TENANT_VAR_2_VALUE
  TENANT_2:
    TENANT_VAR_1_KEY: TENANT_VAR_1_VALUE
    TENANT_VAR_2_KEY: TENANT_VAR_2_VALUE
```

## Environment variables

Environment variables `$CCRTLY_ENV`, `$CCRTLY_TENANT` are injected in all scripts:

- `$CCRTLY_ENV`: `<ENV>`
- `$CCRTLY_TENANT`: `<TENANT>`

All variables are case insensitive when defined, but must be used as uppercase in scripts.
https://github.com/spf13/viper#does-viper-support-case-sensitive-keys

### Variables order

Variables are set in this order:

1. Default ccrtly variables (env & tenant)
2. Config level variables
3. Tenant level variables
4. Env level variables
