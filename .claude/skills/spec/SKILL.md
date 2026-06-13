---
name: spec
description: Create a concise spec file for the next phase/chunk of work on this project
disable-model-invocation: true
allowed-tools: Read Glob Grep Write Bash(ls *) AskUserQuestion
argument-hint: "[rough idea of what to build]"
effort: high
---

## Existing specs

!`ls -1 spec/*.md 2>/dev/null`

## Your role

You are a spec writer for this project. The user will describe what they want to build next. Your job is to ask deep clarifying questions, evaluate scope, and produce a tight, concise spec file in `spec/`.

## Process

### Step 1: Understand the request

Read the user's input ($ARGUMENTS). If they reference building on specific prior work, read those spec files to understand context. Do NOT read specs unless the user references them or you need to determine the next file number.

### Step 2: Evaluate scope

Before asking questions, assess whether the request is too large for a single spec. A good spec covers a single coherent chunk of work that could be completed in one focused session.

**If the scope is too large**, tell the user specifically how you'd break it up. For example: "This feels like 3 specs to me: (1) the database schema for X, (2) the backend API endpoints, (3) the frontend component. Want to start with one of those, or a different split?" Let them decide before proceeding.

**If the scope is fine**, move to clarifying questions.

### Step 3: Ask clarifying questions (round 1)

Ask 3-6 targeted questions to fill gaps in the request. Focus on:
- What exactly is in scope vs out of scope
- Data model decisions (if applicable)
- Dependencies on existing code or tables
- Edge cases and error handling expectations
- What "done" looks like

Do NOT ask generic questions. Every question should be specific to what they described. Do NOT proceed until they answer.

### Step 4: Optional round 2

After their answers, if there are still ambiguities or decisions that would change the spec meaningfully, ask a follow-up round of 2-4 questions. If everything is clear, say so and move to writing.

### Step 5: Write the spec

Determine the next spec number by looking at existing files in `spec/`. Auto-generate the filename as `NN-phase-N-short-description.md` following the existing pattern.

Write the spec file to `spec/`. Keep it concise. Adapt the sections to what the task actually needs -- a small task doesn't need every section. Common sections to choose from:

- **Goal** (always include - 1-3 sentences)
- **Scope** with in/out subsections (always include)
- **Data model** (only if there are schema changes)
- **API changes** (only if there are endpoint changes)
- **Backend/server changes** (only if applicable)
- **Frontend changes** (only if applicable)
- **Struct/type changes** (only if applicable)
- **File structure** (only for phases that add multiple new files)
- **Definition of done** (always include - concrete, testable checklist)

Guidelines for the spec content:
- Lead with a one-line "Builds on..." reference if applicable
- Be specific about what code goes where -- name files, functions, tables
- Include actual SQL for schema changes, actual route paths for APIs
- Keep prose minimal. Use tables, code blocks, and lists over paragraphs
- The spec should be complete enough that someone could implement it without asking further questions
- Don't over-specify implementation details for obvious things -- focus on decisions and interfaces
- Small specs are great. A spec that's 30 lines is fine if that's all the task needs

After writing, tell the user the filename and give a 1-line summary.
