# Pilikino

Pilikino is a Swiss Army knife for note taking apps. It's designed to allow me to operate on my note collection quickly and easily and particularly to facilitate migration between apps without loss of fidelity.

## Features

Pilikino can open a note database and extract individual notes or attachments from them. It can also transfer databases between various formats.

### Database formats supported

- Read/Write - Directory of Markdown files
- Read-only - Joplin Export (JEX) files

### Markdown features supported

- **Links** - when transferring notes, links between notes are automatically updated to use the target format's note linking formula. For example, Joplin `[link](://note_id)` links will be rewritten to `[link](Filename.md)` when writing to a directory of Markdown files.
- **Images** - same as above, when transferring notes, the images are exported into a separate directory and the references are updated.
- **Attachments** - files which are linked to by notes but are not markdown files are exported as well
- **Tables, MathJAX, code blocks** - these are passed through without destroying the formatting.
- **Timestamps of notes** - the modification date of notes is preserved when transferring between databases.

## Future Work

- Should add support for note creation time in addition to modification time.
- Ability to specify a markdown configuration for a database, and be able to convert between Markdown representations.
- Refactor: The `notedb.Note` interface would be more ergonomic to use if I built helper methods similar to `notedb.WriteAST`.
- Refactor: The `Node.IsNote` and `NoteInfo.IsNote` methods should disappear. Notes should be all `.md` files in the virtual filesystem, automatically.

#### Old Version

The source code for the old Pilikino is kept in the [v1 branch](https://github.com/CGamesPlay/pilikino/tree/v1). It's an entirely different application, which I don't intend to use any more.