// utils.js
function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/[&<>]/g, function(m) {
        if (m === '&') return '&amp;';
        if (m === '<') return '&lt;';
        if (m === '>') return '&gt;';
        return m;
    });
}

function showToast(message, isSuccess) {
    var toast = document.getElementById('toast');
    toast.textContent = message;
    toast.classList.remove('success', 'error');
    toast.classList.add(isSuccess ? 'success' : 'error');
    toast.classList.add('show');
    setTimeout(function() {
        toast.classList.remove('show');
    }, 2800);
}

function currentUserType() {
    return document.body?.dataset?.userType || 'normal';
}

function isAdminUser() {
    return currentUserType() === 'admin';
}

function canEditCourseSettings() {
    return currentUserType() === 'admin' || currentUserType() === 'head_ta';
}

function applyRoleVisibility() {
    document.querySelectorAll('[data-admin-only]').forEach(function(el) {
        el.classList.toggle('hidden-by-role', !isAdminUser());
    });
    document.querySelectorAll('[data-settings-only]').forEach(function(el) {
        el.classList.toggle('hidden-by-role', !canEditCourseSettings());
    });
}

function isValidCourseShortName(value) {
    return /^[A-Za-z0-9_]+$/.test(value || '');
}

function toEnglishDigits(value) {
    const persian = '۰۱۲۳۴۵۶۷۸۹';
    const arabic = '٠١٢٣٤٥٦٧٨٩';
    return String(value || '').replace(/[۰-۹٠-٩]/g, function(ch) {
        const p = persian.indexOf(ch);
        if (p > -1) return String(p);
        const a = arabic.indexOf(ch);
        return a > -1 ? String(a) : ch;
    });
}

function toPersianDigits(value) {
    const digits = '۰۱۲۳۴۵۶۷۸۹';
    return String(value || '').replace(/\d/g, function(d) {
        return digits[Number(d)];
    });
}

function pad2(value) {
    return String(value).padStart(2, '0');
}

function gregorianToJalali(gy, gm, gd) {
    const gdm = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334];
    let jy = gy <= 1600 ? 0 : 979;
    gy -= gy <= 1600 ? 621 : 1600;
    const gy2 = gm > 2 ? gy + 1 : gy;
    let days = (365 * gy) + Math.floor((gy2 + 3) / 4) - Math.floor((gy2 + 99) / 100) + Math.floor((gy2 + 399) / 400) - 80 + gd + gdm[gm - 1];
    jy += 33 * Math.floor(days / 12053);
    days %= 12053;
    jy += 4 * Math.floor(days / 1461);
    days %= 1461;
    if (days > 365) {
        jy += Math.floor((days - 1) / 365);
        days = (days - 1) % 365;
    }
    const jm = days < 186 ? 1 + Math.floor(days / 31) : 7 + Math.floor((days - 186) / 30);
    const jd = 1 + (days < 186 ? days % 31 : (days - 186) % 30);
    return { jy: jy, jm: jm, jd: jd };
}

function jalaliToGregorian(jy, jm, jd) {
    jy += 1595;
    let days = -355668 + (365 * jy) + (Math.floor(jy / 33) * 8) + Math.floor(((jy % 33) + 3) / 4) + jd + (jm < 7 ? (jm - 1) * 31 : ((jm - 7) * 30) + 186);
    let gy = 400 * Math.floor(days / 146097);
    days %= 146097;
    if (days > 36524) {
        gy += 100 * Math.floor(--days / 36524);
        days %= 36524;
        if (days >= 365) days++;
    }
    gy += 4 * Math.floor(days / 1461);
    days %= 1461;
    if (days > 365) {
        gy += Math.floor((days - 1) / 365);
        days = (days - 1) % 365;
    }
    let gd = days + 1;
    const leap = (gy % 4 === 0 && gy % 100 !== 0) || (gy % 400 === 0);
    const months = [0, 31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
    let gm = 1;
    while (gm <= 12 && gd > months[gm]) {
        gd -= months[gm];
        gm++;
    }
    return { gy: gy, gm: gm, gd: gd };
}

function getCurrentJalaliYear() {
    const now = new Date();
    return gregorianToJalali(now.getFullYear(), now.getMonth() + 1, now.getDate()).jy;
}

function formatJalaliDateTime(value) {
    if (!value) return '';
    const date = value instanceof Date ? value : new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    const j = gregorianToJalali(date.getFullYear(), date.getMonth() + 1, date.getDate());
    const result = `${j.jy}/${pad2(j.jm)}/${pad2(j.jd)} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`;
    return toPersianDigits(result);
}

function parseJalaliDateTime(value) {
    const normalized = toEnglishDigits(value).trim();
    if (!normalized) return '';
    const match = normalized.match(/^(\d{4})[\/\-](\d{1,2})[\/\-](\d{1,2})(?:\s+(\d{1,2}):(\d{1,2}))?$/);
    if (!match) return null;
    const jy = Number(match[1]);
    const jm = Number(match[2]);
    const jd = Number(match[3]);
    const hour = Number(match[4] || 0);
    const minute = Number(match[5] || 0);
    if (jm < 1 || jm > 12 || jd < 1 || jd > 31 || hour > 23 || minute > 59) return null;
    const g = jalaliToGregorian(jy, jm, jd);
    return `${g.gy}-${pad2(g.gm)}-${pad2(g.gd)}T${pad2(hour)}:${pad2(minute)}`;
}

function normalizeJalaliInputs(root) {
    (root || document).querySelectorAll('.jalali-datetime').forEach(function(input) {
        if (input.dataset.jalaliBound === 'true') return;
        input.dataset.jalaliBound = 'true';
        input.addEventListener('blur', function() {
            const parsed = parseJalaliDateTime(input.value);
            if (input.value.trim() && !parsed) {
                input.classList.add('input-error');
                return;
            }
            input.classList.remove('input-error');
            if (parsed) input.value = formatJalaliDateTime(parsed);
        });
    });
}

function bindFileDropzone(input) {
    if (!input || input.dataset.dropzoneBound === 'true') return;
    input.dataset.dropzoneBound = 'true';
    input.classList.add('native-file-input');

    const box = document.createElement('div');
    box.className = 'file-dropzone';
    box.tabIndex = 0;
    box.setAttribute('role', 'button');
    box.innerHTML = `
        <span class="file-dropzone-icon">📎</span>
        <span class="file-dropzone-name" data-empty="true">انتخاب فایل یا بکشید اینجا</span>
    `;
    input.insertAdjacentElement('afterend', box);

    const choose = function() { input.click(); };
    box.addEventListener('click', choose);
    box.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            choose();
        }
    });
    ['dragenter', 'dragover'].forEach(function(eventName) {
        box.addEventListener(eventName, function(e) {
            e.preventDefault();
            box.classList.add('is-dragover');
        });
    });
    ['dragleave', 'drop'].forEach(function(eventName) {
        box.addEventListener(eventName, function(e) {
            e.preventDefault();
            box.classList.remove('is-dragover');
        });
    });
    box.addEventListener('drop', function(e) {
        const files = e.dataTransfer && e.dataTransfer.files;
        if (!files || !files.length) return;
        try {
            const transfer = new DataTransfer();
            Array.from(files).forEach(function(file) { transfer.items.add(file); });
            input.files = transfer.files;
        } catch (_) {
            return;
        }
        input.dispatchEvent(new Event('change', { bubbles: true }));
    });

    const updateBox = function() {
        const name = box.querySelector('.file-dropzone-name');
        if (!name) return;
        const files = Array.from(input.files || []);
        if (files.length) {
            const label = files.map(function(f) { return f.name; }).join('، ');
            if (name.textContent !== label) name.textContent = label;
            if (name.dataset.empty !== 'false') name.dataset.empty = 'false';
            if (!box.classList.contains('has-file')) box.classList.add('has-file');
        } else {
            const label = 'انتخاب فایل یا بکشید اینجا';
            if (name.textContent !== label) name.textContent = label;
            if (name.dataset.empty !== 'true') name.dataset.empty = 'true';
            if (box.classList.contains('has-file')) box.classList.remove('has-file');
        }
    };
    input.addEventListener('change', updateBox);
    // Run once for initial state
    updateBox();

    // Listen for form reset on the closest form only (avoids a global listener)
    const form = input.form;
    if (form && !form.dataset.dropzoneResetBound) {
        form.dataset.dropzoneResetBound = 'true';
        form.addEventListener('reset', function() {
            setTimeout(function() {
                form.querySelectorAll('input[type="file"][data-dropzone-bound="true"]').forEach(function(inp) {
                    inp.dispatchEvent(new Event('change', { bubbles: true }));
                });
            }, 0);
        });
    }
}

function initFileDropzones(root) {
    const scope = root || document;
    scope.querySelectorAll('input[type="file"]').forEach(bindFileDropzone);
}

// Backwards-compat shim: explicit refresh by re-dispatching change events
function refreshFileDropzones(root) {
    (root || document).querySelectorAll('input[type="file"][data-dropzone-bound="true"]').forEach(function(input) {
        input.dispatchEvent(new Event('change', { bubbles: true }));
    });
}
