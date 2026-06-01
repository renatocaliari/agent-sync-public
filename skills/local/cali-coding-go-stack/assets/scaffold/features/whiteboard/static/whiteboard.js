document.addEventListener('DOMContentLoaded', function() {
    const container = document.getElementById('fabric-canvas').parentElement;
    if (!container) return;
    const canvas = new fabric.Canvas('fabric-canvas', { isDrawingMode: false, selection: true, width: container.clientWidth || 800, height: container.clientHeight || 600 });
    const tabId = sessionStorage.getItem('wb') || sessionStorage.setItem('wb', Math.random().toString(36).substr(2,6)) || sessionStorage.getItem('wb');
    let tool = 'select', color = '#000', brush = 2, drawing = false, start = null, temp = null, remote = false;
    const states = new Map();
    new ResizeObserver(() => { canvas.setWidth(container.clientWidth); canvas.setHeight(container.clientHeight); canvas.renderAll(); }).observe(container);
    document.querySelectorAll('[data-on\\:click*="$tool"]').forEach(b => b.addEventListener('click', () => {
        const m = b.getAttribute('data-on:click').match(/\$tool='(\w+)'/);
        if (m) { tool = m[1]; canvas.isDrawingMode = tool === 'pencil'; canvas.selection = tool === 'select';
            if (tool === 'pencil') { canvas.freeDrawingBrush = new fabric.PencilBrush(canvas); canvas.freeDrawingBrush.color = color; canvas.freeDrawingBrush.width = brush; }
            document.querySelectorAll('[data-on\\:click*="$tool"]').forEach(x => x.classList.remove('btn-active')); b.classList.add('btn-active');
        }
    }));
    document.querySelectorAll('[data-on\\:click*="$color"]').forEach(b => b.addEventListener('click', () => {
        const m = b.getAttribute('data-on:click').match(/\$color='(#\w+)'/);
        if (m) { color = m[1]; if (canvas.isDrawingMode) canvas.freeDrawingBrush.color = color;
            document.querySelectorAll('[data-on\\:click*="$color"]').forEach(x => x.classList.remove('btn-active')); b.classList.add('btn-active');
        }
    }));
    function send(d, persist = true) { fetch('/whiteboard/drawing?tabId=' + tabId, { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ id: d.id || 'obj_'+Date.now(), tool: d.type || tool, color: color, delta: d, _live: !persist, userId: tabId }) }); }
    canvas.on('mouse:down', o => { if (tool === 'select' || tool === 'pencil') return; drawing = true; start = canvas.getPointer(o.e);
        const c = { left: start.x, top: start.y, selectable: false, stroke: color, strokeWidth: brush, fill: 'transparent', id: 's'+Date.now() };
        if (tool === 'rect') temp = new fabric.Rect({ ...c, width: 0, height: 0 });
        else if (tool === 'circle') temp = new fabric.Circle({ ...c, radius: 0 });
        if (temp) { canvas.add(temp); canvas.renderAll(); }
    });
    canvas.on('mouse:move', o => { if (!drawing || !temp) return; const p = canvas.getPointer(o.e);
        if (tool === 'rect') temp.set({ width: Math.abs(p.x - start.x), height: Math.abs(p.y - start.y), left: Math.min(p.x, start.x), top: Math.min(p.y, start.y) });
        else if (tool === 'circle') { const r = Math.sqrt(Math.pow(p.x - start.x, 2) + Math.pow(p.y - start.y, 2)) / 2; temp.set({ radius: r, left: Math.min(p.x, start.x), top: Math.min(p.y, start.y) }); }
        canvas.renderAll();
    });
    canvas.on('mouse:up', () => { if (!drawing || !temp) return; drawing = false; temp.set({ selectable: true });
        states.set(temp.id, temp.toObject(['id'])); send(temp.toObject(['id', 'type', 'left', 'top', 'width', 'height', 'radius', 'stroke'])); temp = null;
    });
    canvas.on('path:created', e => { if (remote) return; const p = e.path; p.set({ id: 'p'+Date.now(), fill: 'transparent', stroke: color, strokeWidth: brush, strokeLineCap: 'round', strokeLineJoin: 'round' });
        states.set(p.id, p.toObject(['id', 'path'])); send(p.toObject(['id', 'type', 'path', 'stroke', 'strokeWidth']));
    });
    canvas.on('object:modified', e => { if (remote) return; const o = e.target, orig = states.get(o.id) || {}, d = { id: o.id, type: o.type };
        ['left','top','width','height','scaleX','scaleY','angle'].forEach(k => { if (o[k] !== undefined && o[k] !== orig[k]) d[k] = o[k]; });
        if (Object.keys(d).length > 2) { send(d); states.set(o.id, o.toObject(['id'])); }
    });
    document.getElementById('upload-btn').addEventListener('click', () => document.getElementById('file-input').click());
    document.getElementById('file-input').addEventListener('change', e => { const f = e.target.files[0]; if (f) {
        const r = new FileReader(); r.onload = ev => { fabric.Image.fromURL(ev.target.result, img => {
            img.set({ left: canvas.width/2-img.width/4, top: canvas.height/2-img.height/4, scaleX: 0.5, scaleY: 0.5, selectable: true, id: 'i'+Date.now() });
            canvas.add(img); canvas.renderAll(); send(img.toObject(['id', 'src', 'left', 'top', 'scaleX', 'scaleY']));
        }); }; r.readAsDataURL(f); } e.target.value = '';
    });
    document.getElementById('clear-btn').addEventListener('click', () => { if (confirm('Clear?')) { fetch('/whiteboard/drawings?tabId=' + tabId, { method: 'DELETE' }); canvas.clear(); states.clear(); } });
    document.addEventListener('paste', e => { e.preventDefault(); for (let i of e.clipboardData.items) { if (i.type.indexOf('image') !== -1) {
        const b = i.getAsFile(), r = new FileReader(); r.onload = ev => { fabric.Image.fromURL(ev.target.result, img => { img.set({ left: canvas.width/2-img.width/4, top: canvas.height/2-img.height/4, scaleX: 0.5, id: 'i'+Date.now() }); canvas.add(img); canvas.renderAll(); send(img.toObject(['id', 'src', 'left', 'top'])); }); }; r.readAsDataURL(b); return; }
    }});
    document.addEventListener('keydown', e => { if (e.target.tagName === 'INPUT') return; const a = canvas.getActiveObject();
        if ((e.key === 'Delete' || e.key === 'Backspace') && a) { fetch('/whiteboard/object/' + a.id + '?tabId=' + tabId, { method: 'DELETE' }); canvas.remove(a); canvas.discardActiveObject(); canvas.renderAll(); }
    });
    window.addEventListener('wb:canvas_init', e => { const d = e.detail; if (d.objects) { remote = true; canvas.loadFromJSON(d, () => { canvas.renderAll(); remote = false; canvas.getObjects().forEach(o => states.set(o.id, o.toObject(['id']))); }); }});
    window.addEventListener('wb:canvas_delta', e => { const d = e.detail; if (d.userId === tabId) return; remote = true;
        if (d.type === 'object') { const f = d.fabricData || d.delta; if (!f) { remote = false; return; }
            if (f.type === 'path') { f.fill = 'transparent'; f.strokeLineCap = 'round'; f.strokeLineJoin = 'round'; }
            const ex = canvas.getObjects().find(o => o.id === f.id);
            if (ex) { ['left','top','width','height','scaleX','scaleY','angle','stroke','strokeWidth'].forEach(k => { if (f[k] !== undefined) ex.set(k, f[k]); }); ex.setCoords(); canvas.renderAll(); }
            else { fabric.util.enlivenObjects([f], objs => { objs.forEach(o => { o.set('id', f.id); canvas.add(o); states.set(o.id, o.toObject(['id'])); }); canvas.renderAll(); remote = false; }); return; }
        } else if (d.type === 'clear') { canvas.clear(); states.clear(); } remote = false;
    });
    const curs = {}; let myId = null;
    window.addEventListener('wb:cursor_init', e => { const d = e.detail; if (d.myCursor) myId = d.myCursor.userId; });
    window.addEventListener('wb:cursor_update', e => { const c = e.detail; if (c.userId === myId) return;
        let el = curs[c.userId]; if (!el) { el = document.createElement('div'); el.style.cssText = 'position:absolute;pointer-events:none;z-index:50;';
            el.innerHTML = '<svg width="24" height="24" style="color:'+c.color+'"><path fill="currentColor" d="M4 4l16 8-7 2-2 7z"/></svg><span style="margin-left:20px;padding:2px 6px;background:'+c.color+';color:white;font-size:11px;border-radius:4px">'+c.name+'</span>';
            document.getElementById('cursors-layer').appendChild(el); curs[c.userId] = el; }
        el.style.left = c.x + 'px'; el.style.top = c.y + 'px';
    });
    canvas.on('mouse:move', e => { const p = canvas.getPointer(e.e); fetch('/whiteboard/cursor?tabId=' + tabId, { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ x: p.x, y: p.y }) }); });
});
