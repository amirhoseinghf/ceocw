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