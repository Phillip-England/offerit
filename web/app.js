const form = document.querySelector('#letter-form');
const frame = document.querySelector('#pdf-preview');
const status = document.querySelector('#status');
const download = document.querySelector('#download');
const empty = document.querySelector('#preview-empty');
const buttons = [...document.querySelectorAll('.template')];
let previewURL = null;
let timer = null;
let version = 0;

const templates = {
  general: {
    headline: 'A website that works for your business.',
    opening: 'Hello, I’m Phillip England. I build websites and practical software for small businesses. I wanted to introduce myself and share how I might help your business online.',
    opportunity: 'A clear website can show people what you do, answer common questions, and make it easy to find your hours, location, and services. I would start by learning what matters most to your business.',
    offer: 'I design and build websites, and I can create custom software for the work that happens behind the scenes. A starter website begins at $300, with hosting priced separately. I can also scope a more tailored project with you.',
    nextStep: 'Scan the code below or visit englandsoftware.com to see my work and get in touch.'
  },
  built: {
    headline: 'I built a website for your business.',
    opening: 'Hello, I’m Phillip England. I put together a live website concept for your business so you can see what a polished online presence could look like. You can explore it at the link below.',
    opportunity: '',
    offer: '',
    nextStep: 'If you would like to purchase the site, text or call me at (918) 500-8396. I would be happy to walk you through it.'
  }
};

function data() { return Object.fromEntries(new FormData(form).entries()); }
function updateCounts() {
  for (const name of ['opening', 'opportunity', 'offer', 'nextStep']) {
    const field = form.elements[name];
    document.getElementById(name + '-count').textContent = `${field.value.length}/${field.maxLength}`;
  }
}
function save() { localStorage.setItem('offerit-draft-v2', JSON.stringify(data())); }
function applyKind(kind) {
  form.elements.kind.value = kind;
  const built = kind === 'built';
  document.querySelector('#general-fields').hidden = built;
  document.querySelector('#built-fields').hidden = !built;
  for (const name of ['opportunity', 'offer']) form.elements[name].disabled = built;
  for (const name of ['demoUrl', 'removalDate']) form.elements[name].disabled = !built;
  form.elements.demoUrl.required = built;
  buttons.forEach(button => {
    const active = button.dataset.template === kind;
    button.classList.toggle('active', active);
    button.setAttribute('aria-pressed', active);
  });
  document.querySelector('#message-note').textContent = built
    ? 'Share their live demo and make the price and purchase step clear.'
    : 'Make the introduction personal, then invite them to your website.';
  document.querySelector('#footer-note').textContent = built
    ? 'The QR code points to the live demo. Your phone number appears beside it for texts or calls.'
    : 'The QR code points to englandsoftware.com. The upper right stays open for your business card.';
}
function setTemplate(kind) {
  if (kind === form.elements.kind.value) return;
  for (const [key, value] of Object.entries(templates[kind])) form.elements[key].value = value;
  applyKind(kind);
  changed();
}
function changed() {
  updateCounts();
  save();
  clearTimeout(timer);
  timer = setTimeout(preview, 450);
}
async function makePDF(downloadFile = false) {
  const response = await fetch('/api/pdf' + (downloadFile ? '?download=1' : ''), {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data())
  });
  if (!response.ok) throw new Error((await response.text()).trim());
  return { blob: await response.blob(), disposition: response.headers.get('Content-Disposition') || '' };
}
async function preview() {
  const current = ++version;
  if (!form.elements.business.value.trim() || (form.elements.kind.value === 'built' && !form.elements.demoUrl.value.trim())) {
    frame.removeAttribute('src');
    empty.style.display = 'grid';
    empty.textContent = !form.elements.business.value.trim() ? 'Enter a business name to preview your letter.' : 'Enter the live demo URL to preview this letter.';
    status.textContent = !form.elements.business.value.trim() ? 'Add a business name to begin.' : 'Add the live demo URL to continue.';
    return;
  }
  status.textContent = 'Updating preview…';
  try {
    const { blob } = await makePDF();
    if (current !== version) return;
    const url = URL.createObjectURL(blob);
    frame.src = url + '#toolbar=0&navpanes=0';
    empty.style.display = 'none';
    if (previewURL) setTimeout(() => URL.revokeObjectURL(previewURL), 2000);
    previewURL = url;
    status.textContent = 'Ready to print or share.';
  } catch (error) {
    if (current !== version) return;
    const message = error.message.replace(/<[^>]*>/g, '') || 'Preview unavailable.';
    frame.removeAttribute('src');
    empty.style.display = 'grid';
    empty.textContent = message;
    status.textContent = message;
  }
}
form.addEventListener('input', changed);
buttons.forEach(button => button.addEventListener('click', () => setTemplate(button.dataset.template)));
download.addEventListener('click', async () => {
  if (!form.reportValidity()) return;
  download.disabled = true;
  status.textContent = 'Preparing your PDF…';
  try {
    const { blob, disposition } = await makePDF(true);
    const name = /filename="([^"]+)"/.exec(disposition)?.[1] || 'offer-letter.pdf';
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url; link.download = name; document.body.append(link); link.click(); link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 5000);
    status.textContent = 'PDF downloaded.';
  } catch (error) {
    status.textContent = error.message.replace(/<[^>]*>/g, '') || 'Download failed.';
  } finally { download.disabled = false; }
});

let saved = null;
try { saved = JSON.parse(localStorage.getItem('offerit-draft-v2') || 'null'); } catch {}
if (saved) {
  for (const [key, value] of Object.entries(saved)) if (form.elements[key]) form.elements[key].value = value;
  applyKind(saved.kind === 'built' ? 'built' : 'general');
} else {
  for (const [key, value] of Object.entries(templates.general)) form.elements[key].value = value;
  applyKind('general');
}
updateCounts();
preview();
