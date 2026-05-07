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

function bindFileDropzone(input) {
    if (!input || input.dataset.dropzoneBound === 'true') return;
    input.dataset.dropzoneBound = 'true';
    input.classList.add('native-file-input');

    var box = document.createElement('div');
    box.className = 'file-dropzone';
    box.tabIndex = 0;
    box.setAttribute('role', 'button');
    box.innerHTML = '<span class="file-dropzone-icon">📎</span><span class="file-dropzone-name" data-empty="true">انتخاب فایل یا بکشید اینجا</span>';
    input.insertAdjacentElement('afterend', box);

    var choose = function() { input.click(); };
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
        var files = e.dataTransfer && e.dataTransfer.files;
        if (!files || !files.length) return;
        try {
            var transfer = new DataTransfer();
            Array.from(files).forEach(function(file) { transfer.items.add(file); });
            input.files = transfer.files;
        } catch (_) {
            return;
        }
        input.dispatchEvent(new Event('change', { bubbles: true }));
    });

    var updateBox = function() {
        var name = box.querySelector('.file-dropzone-name');
        if (!name) return;
        var files = Array.from(input.files || []);
        if (files.length) {
            var label = files.map(function(file) { return file.name; }).join('، ');
            name.textContent = label;
            name.dataset.empty = 'false';
            box.classList.add('has-file');
        } else {
            name.textContent = 'انتخاب فایل یا بکشید اینجا';
            name.dataset.empty = 'true';
            box.classList.remove('has-file');
        }
    };
    input.addEventListener('change', updateBox);
    updateBox();

    var form = input.form;
    if (form && !form.dataset.dropzoneResetBound) {
        form.dataset.dropzoneResetBound = 'true';
        form.addEventListener('reset', function() {
            setTimeout(function() {
                form.querySelectorAll('input[type="file"][data-dropzone-bound="true"]').forEach(function(element) {
                    element.dispatchEvent(new Event('change', { bubbles: true }));
                });
            }, 0);
        });
    }
}

function refreshFileDropzones(root) {
    (root || document).querySelectorAll('input[type="file"][data-dropzone-bound="true"]').forEach(function(input) {
        input.dispatchEvent(new Event('change', { bubbles: true }));
    });
}