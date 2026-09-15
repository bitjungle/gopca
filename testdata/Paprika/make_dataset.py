"""Rebuild paprika.xlsx from the two published tables.

The workbook this produces is a compilation of material from Fiamegos et al.
(2021), which is published under CC BY-NC-ND 4.0. That licence permits sharing
the article verbatim but not sharing adapted material, and combining two tables
into a restructured workbook is plausibly adaptation rather than the mere change
of format that the licence explicitly allows. So the repository carries this
script instead of the workbook: anyone with the two source documents can
reproduce the dataset byte for byte, and no publisher content is redistributed.

Inputs, neither of which is in the repository:

  1-s2.0-S0956713520304126-mmc1.docx   Table S1, the elemental mass fractions.
                                       Download from the article's Appendix A:
                                       https://doi.org/10.1016/j.foodcont.2020.107496

  the article PDF                      Table 1, the sample descriptions. Looked
                                       for in docs/references/ by default, which
                                       is local-only.

Output: paprika.xlsx, three sheets -- data, sample_info, meta.

No third-party packages. A .docx and an .xlsx are both zip archives of XML, and
the standard library reads and writes both; the only external requirement is
pdftotext (poppler) for the article PDF.

Run from this directory:  python make_dataset.py
"""

import argparse
import re
import subprocess
import sys
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

W = '{http://schemas.openxmlformats.org/wordprocessingml/2006/main}'

ELEMENTS = ['P', 'Cl', 'S', 'K', 'Ca', 'Cr', 'Mn', 'Fe',
            'Ni', 'Cu', 'Zn', 'Br', 'Rb', 'Sr', 'Ba']

# Five elements are reported in g/kg and ten in mg/kg. Keeping the units in a
# second header row is how the published table presents them, and it is also
# what makes this a useful import test: a naive reader takes the units row for
# a sample and ends up with 68 rows, one of them text in every column.
UNITS = ['(g kg-1)'] * 5 + ['(mg kg-1)'] * 10

# Counts stated in the paper, used below to check the parse rather than to
# trust it. A silent mis-parse that produced 66 or 68 samples would otherwise
# look exactly like a successful run.
EXPECTED_TOTAL = 67
EXPECTED_GROUPS = {'LV': 33, 'SNVL': 7, 'Rest': 27}

INFO_COLUMNS = ['Sample code', 'Type of paprika', 'Smoked', 'Group in the study',
                'Country of origin', 'Country of purchase', 'Year of purchase']

META = [
    ('Article:', 'https://doi.org/10.1016/j.foodcont.2020.107496'),
    ('Groups:', 'Pimentón de La Vera (LV), Spanish not from La Vera (SNLV), '
                'not Spanish or of unknown origin (Rest)'),
    ('Analytical method:', 'Energy Dispersive X-ray Fluorescence (ED-XRF)'),
    ('Instrument:', 'Epsilon 5 ED-XRF spectrometer'),
    ('< LoQ', 'Below the limit of quantification'),
]


# --- reading Table S1 out of the supplementary .docx -------------------------

def read_table_s1(path):
    """Return [[sample_code, group, *15 mass fractions], ...] from the .docx.

    The table is parsed through its w:tbl structure rather than by flattening
    the document to text, because the group labels (LV, SNVL, Rest) sit in rows
    of their own and flattening loses the boundary between a label and the
    samples beneath it.
    """
    with zipfile.ZipFile(path) as z:
        doc = ET.fromstring(z.read('word/document.xml'))

    table = doc.find(f'.//{W}tbl')
    if table is None:
        sys.exit(f'{path}: no table found; is this the right supplementary file?')

    rows = []
    for tr in table.findall(f'{W}tr'):
        cells = []
        for tc in tr.findall(f'{W}tc'):
            text = ''.join(t.text or '' for t in tc.iter(f'{W}t'))
            # The published table writes units as "g kg -1" with stray spaces,
            # and "< LoQ" appears both with and without the space.
            cells.append(re.sub(r'\s+', ' ', text).strip())
        rows.append(cells)

    samples, group = [], None
    for cells in rows:
        if not cells:
            continue
        first = cells[0]
        if first in EXPECTED_GROUPS:
            # A row carrying only a group label starts a new block.
            group = first
            continue
        if not re.fullmatch(r'PAPR\d{4}', first):
            continue  # header, or the Median/Mean/STD/RSD summary rows
        if group is None:
            sys.exit(f'{first} appears before any group label; parse is wrong')
        values = [c for c in cells[1:] if c != ''][:len(ELEMENTS)]
        if len(values) != len(ELEMENTS):
            sys.exit(f'{first}: found {len(values)} values, expected {len(ELEMENTS)}')
        samples.append([first, group] + values)

    return samples


# --- reading Table 1 out of the article PDF ----------------------------------

def read_table_1(pdf_path, page=2):
    """Return [[sample_code, type, smoked, group, origin, purchase, year], ...].

    pdftotext -layout preserves the column alignment, so the fields are sliced
    at the offsets of the header labels. Splitting on whitespace would not work:
    several values contain spaces ("Hungary & Spain", "March 2018").
    """
    try:
        out = subprocess.run(
            ['pdftotext', '-f', str(page), '-l', str(page), '-layout', str(pdf_path), '-'],
            capture_output=True, text=True, check=True).stdout
    except FileNotFoundError:
        sys.exit('pdftotext not found. Install poppler (brew install poppler).')
    except subprocess.CalledProcessError as exc:
        sys.exit(f'pdftotext failed on {pdf_path}: {exc}')

    header = next((ln for ln in out.splitlines() if 'Sample code' in ln), None)
    if header is None:
        sys.exit(f'{pdf_path}: no Table 1 header on page {page}')

    starts = []
    for name in INFO_COLUMNS:
        at = header.find(name)
        if at < 0:
            sys.exit(f'{pdf_path}: column {name!r} missing from Table 1 header')
        starts.append(at)
    bounds = list(zip(starts, starts[1:] + [len(header) + 200]))

    rows = []
    for line in out.splitlines():
        if not re.match(r'\s+PAPR\d{4}', line):
            continue
        rows.append([line[a:b].strip() for a, b in bounds])
    return rows


# --- writing the workbook ----------------------------------------------------

def col_ref(index):
    """0 -> A, 25 -> Z, 26 -> AA."""
    ref = ''
    index += 1
    while index:
        index, rem = divmod(index - 1, 26)
        ref = chr(65 + rem) + ref
    return ref


def escape(text):
    return (text.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;'))


def is_number(value):
    try:
        float(value)
        return True
    except ValueError:
        return False


def sheet_xml(rows):
    """One worksheet. Inline strings, so no shared-string table is needed."""
    out = ['<?xml version="1.0" encoding="UTF-8" standalone="yes"?>',
           '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">',
           '<sheetData>']
    for r, row in enumerate(rows, start=1):
        out.append(f'<row r="{r}">')
        for c, value in enumerate(row):
            if value == '':
                continue
            ref = f'{col_ref(c)}{r}'
            if is_number(value):
                out.append(f'<c r="{ref}"><v>{value}</v></c>')
            else:
                out.append(f'<c r="{ref}" t="inlineStr"><is><t xml:space="preserve">'
                           f'{escape(value)}</t></is></c>')
        out.append('</row>')
    out.append('</sheetData></worksheet>')
    return ''.join(out)


def write_xlsx(path, sheets):
    """sheets: [(name, rows), ...]"""
    types = ['<?xml version="1.0" encoding="UTF-8" standalone="yes"?>',
             '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">',
             '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>',
             '<Default Extension="xml" ContentType="application/xml"/>',
             '<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>']
    books, rels = [], []
    for i, (name, _) in enumerate(sheets, start=1):
        types.append(f'<Override PartName="/xl/worksheets/sheet{i}.xml" '
                     'ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>')
        books.append(f'<sheet name="{escape(name)}" sheetId="{i}" r:id="rId{i}"/>')
        rels.append(f'<Relationship Id="rId{i}" '
                    'Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" '
                    f'Target="worksheets/sheet{i}.xml"/>')
    types.append('</Types>')

    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml', ''.join(types))
        z.writestr('_rels/.rels',
                   '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
                   '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
                   '<Relationship Id="rId1" '
                   'Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" '
                   'Target="xl/workbook.xml"/></Relationships>')
        z.writestr('xl/workbook.xml',
                   '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
                   '<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" '
                   'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">'
                   f'<sheets>{"".join(books)}</sheets></workbook>')
        z.writestr('xl/_rels/workbook.xml.rels',
                   '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
                   '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
                   f'{"".join(rels)}</Relationships>')
        for i, (_, rows) in enumerate(sheets, start=1):
            z.writestr(f'xl/worksheets/sheet{i}.xml', sheet_xml(rows))


# --- checks ------------------------------------------------------------------

def check(samples, info):
    """Fail loudly rather than write a workbook that is quietly wrong."""
    if len(samples) != EXPECTED_TOTAL:
        sys.exit(f'Table S1 gave {len(samples)} samples, expected {EXPECTED_TOTAL}')
    if len(info) != EXPECTED_TOTAL:
        sys.exit(f'Table 1 gave {len(info)} samples, expected {EXPECTED_TOTAL}')

    counts = {}
    for row in samples:
        counts[row[1]] = counts.get(row[1], 0) + 1
    if counts != EXPECTED_GROUPS:
        sys.exit(f'group counts {counts}, expected {EXPECTED_GROUPS}')

    s_codes = {r[0] for r in samples}
    i_codes = {r[0] for r in info}
    if s_codes != i_codes:
        sys.exit(f'sample codes differ between the tables: '
                 f'only in S1 {sorted(s_codes - i_codes)}, '
                 f'only in T1 {sorted(i_codes - s_codes)}')

    # The two tables list the same samples in a different order. The sheets are
    # written in their published orders, so anything joining them must join on
    # the sample code -- which is worth reporting rather than silently fixing.
    misaligned = sum(1 for a, b in zip(samples, info) if a[0] != b[0])
    if misaligned:
        print(f'  note: {misaligned} rows appear in a different order in the two '
              f'tables; join on sample code, not row position')


def main():
    here = Path(__file__).resolve().parent
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument('--docx', type=Path,
                    default=here / '1-s2.0-S0956713520304126-mmc1.docx')
    ap.add_argument('--pdf', type=Path,
                    default=here / '../../docs/references/Fiamegos et al. - 2021 - '
                                   'Authentication of PDO paprika powder by  mva of '
                                   'the elemental fingerprint determined by XRF.pdf')
    ap.add_argument('--out', type=Path, default=here / 'paprika.xlsx')
    args = ap.parse_args()

    for path, what in ((args.docx, 'supplementary .docx'), (args.pdf, 'article PDF')):
        if not path.exists():
            sys.exit(f'{what} not found at {path}\nSee the module docstring for where '
                     f'to obtain it, or pass an explicit path.')

    samples = read_table_s1(args.docx)
    info = read_table_1(args.pdf)
    check(samples, info)

    # The group label is spelled SNVL in Table S1 and SNLV in Table 1, both as
    # published. Neither is corrected here, for the same reason the aluminium
    # alloy file keeps its malformed row: this copy should not disagree with the
    # source. The README records the discrepancy.
    data_rows = [[''] + ['CLASS'] + ELEMENTS,
                 [' '] + [''] + UNITS]
    data_rows += [[row[0], row[1]] + row[2:] for row in samples]

    info_rows = [INFO_COLUMNS] + info
    meta_rows = [list(pair) for pair in META]

    write_xlsx(args.out, [('data', data_rows),
                          ('sample_info', info_rows),
                          ('meta', meta_rows)])
    print(f'wrote {args.out}')
    print(f'  data        {len(data_rows) - 2} samples x {len(ELEMENTS)} elements')
    print(f'  sample_info {len(info_rows) - 1} samples x {len(INFO_COLUMNS)} fields')
    print(f'  meta        {len(meta_rows)} rows')


if __name__ == '__main__':
    main()
