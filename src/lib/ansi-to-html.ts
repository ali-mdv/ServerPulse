import Convert from "ansi-html";

export function ansiToHtml(text: string): string {
  // Remove non-printable control characters but preserve ANSI escape sequences
  // Keep only valid characters and ANSI escape sequences (ESC = \x1B)
  text = text.replace(/[\x00-\x08\x0B-\x0C\x0E-\x1A\x7F-\x9F]/g, "");

  // Use the ansi-html library to convert ANSI codes to HTML
  // Default colors: black background, white text
  return Convert(text);
}
