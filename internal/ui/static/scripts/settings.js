// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only
document.body.addEventListener('htmx:beforeSwap', function(evt) {
    if(evt.detail.xhr.status === 422){
       evt.detail.shouldSwap = true;
       evt.detail.isError = false;
    } else if(evt.detail.xhr.status === 500) {
       evt.detail.shouldSwap = true;
    }
});
