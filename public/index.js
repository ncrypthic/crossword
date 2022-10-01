const { createElement, useState } = React;
const render = ReactDOM.render;
const html = htm.bind(createElement);

function CrosswordApp() {
  const [size, setSize] = useState(10);
  const [crossword, setCrossword] = useState({matrix: [], ans: [], hints: []});
  const [drag, setDrag] = useState(false);
  const [selection, setSelection] = useState([]);
  const [foundWords, setFoundWords] = useState([]);
  const [finished, setFinished] = useState(false);

  const fetchCrossword = async () => {
    const res = await fetch("/crossword", {
      method: "POST",
      body: JSON.stringify({size: parseInt(size, 10)})
    });
    setCrossword(await res.json());
    setFoundWords([]);
    setSelection([]);
    setDrag(false);
  }

  const isSelected = (x, y) => {
    if (foundWords.some(elmt => elmt.some(([row, col]) => row == x && col == y))) {
      return true;
    }
    if(finished) {
      return false;
    }
    if(!drag) return;
    const word = selection.reduce((str, c) => str+=c[2], "")
    if(crossword.ans[word]) {
      setTimeout(() => {
        const hashmap = selection.reduce((aggr, cell) => {
          aggr[cell.toString()] = cell
          return aggr;
        }, {});
        const unique = Object.keys(hashmap).map(k => hashmap[k]);
        setFoundWords(foundWords.concat([unique]));
        setSelection([]);
        if (Object.keys(crossword.ans).every(isSolved)) {
          setFinished(true);
        }
      }, 100);

      return true;
    }
    return selection.some(([row, col]) => row == x && col == y);
  }

  const isSolved = (query) => {
    const words = foundWords.reduce((res, cells) =>
      res.concat([cells.reduce((str, cell) => str+=cell[2], "")]
    ), []);
    return words.some(w => w === query)
  }

  const resetPuzzle = () => {
    setCrossword({matrix: [], ans: [], hints: []});
    setFinished(false);
  }

  return html`
    <div id="content">
      <div>
        <fieldset>
          <field>
            <label>Puzzle Size</label>
            <input type="number" onChange=${(e) => setSize(e.target.value)} value=${size}/>
          </field>
          <button type="button" onClick=${fetchCrossword}>Start</button>
        </fieldset>
      </div>
      <div style=${{display: 'flex', flexDirection: 'row'}}>
        <div className="hint" style=${{flex:  1}}>
          <ol>
            ${Object.keys(crossword.hints).map(k => {
              return html`<li className=${isSolved(k) ? 'solved' : ''}>${crossword.hints[k]}</li>`
            })}
          </ol>
        </div>
        <div style=${{flex: 3, position: 'relative'}}>
          ${crossword.matrix.length > 0 ?
            html`
          <table cellpadding="0" cellspacing="0" className=${finished ? "solved" : null}>
            <thead>
              <tr>
                <th width="20"></th>
                ${crossword.matrix.map((_, i) => {
                  return html`<th width="20">${i+1}</th>`;
                })}
              </tr>
            </thead>
            ${crossword.matrix.map((row, x) => {
              return html`<tr>
                <td>${x+1}</td>
              ${row.map((cell, y) => {
                return html`
                <td onMouseEnter=${e => {
                  setSelection(selection.concat([[x, y, cell]]))
                }} onMouseUp=${e => {
                  setDrag(false);
                  setSelection([]);
                }} onMouseDown=${e => {
                  setDrag(true);
                  setSelection([[x, y, cell]])
                }} className=${isSelected(x, y) ? 'selected': null}>
                  ${cell}
                </td>`
              })}
              </tr>`
            })}
          </table>
          `: null}
          ${finished ? html`<div style=${{fontSize: '16rem', position: 'absolute', top: 0, left: 0}}><span>🎉</span><br/><button type="button" onClick=${resetPuzzle}>Reset</button></div>` : null }
        </div>
      </div>
    </div>
  `;
}

render(html`<${CrosswordApp}/>`, document.getElementById("root"));
