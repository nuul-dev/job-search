#!/usr/bin/env python3
"""
Build a PDF from a resume HTML file using WeasyPrint.

Usage:
  python3 build-pdf.py <input.html> <output.pdf>

If WeasyPrint is missing, install:
  pip install weasyprint
"""
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 3:
        print("Usage: build-pdf.py <input.html> <output.pdf>", file=sys.stderr)
        return 2

    input_path = Path(sys.argv[1])
    output_path = Path(sys.argv[2])

    if not input_path.exists():
        print(f"HTML not found: {input_path}", file=sys.stderr)
        return 1

    try:
        import weasyprint
    except ImportError:
        print("WeasyPrint not installed. Run: pip install weasyprint", file=sys.stderr)
        return 1

    weasyprint.HTML(filename=str(input_path)).write_pdf(str(output_path))
    print(f"Wrote {output_path} ({output_path.stat().st_size // 1024} KB)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
