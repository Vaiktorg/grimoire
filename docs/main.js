
function toggleEditor() {
    let editor = document.querySelector('#editor');

    if (editor.hasAttribute('hidden')) {
        editor.removeAttribute('hidden');
    } else {
        editor.setAttribute('hidden','true');
    }
}