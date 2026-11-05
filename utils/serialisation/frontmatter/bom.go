/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

// utf8BOM is the Unicode byte order mark character sometimes emitted at the
// start of UTF-8 encoded text files.
//
// Although a BOM is not required for UTF-8, some editors, operating systems,
// generators, or copy/paste workflows still prepend it. In practice this means
// Markdown documents, README files, and frontmatter-bearing content can begin
// with this character before the opening delimiter, for example before `---`.
//
// Parsers that match delimiters on the first line therefore need to tolerate it,
// otherwise a valid frontmatter block may be missed even though the visible text
// looks correct.
const utf8BOM = "\uFEFF"
