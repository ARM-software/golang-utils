<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Migration guides

These guides describe transitions between supported APIs and import paths:

- [Replacing ozzo `validation.Required`](ozzo-required.md): preserve pre-v4.4
  zero-valued struct semantics with `utils/validation.IsRequired`.
- [Refactoring Go modules without breaking imports](refactoring-go-modules.md):
  split modules, preserve old paths, and verify compatibility.
- [Serialisation packages](serialisation.md): new import paths and existing
  compatible imports.
