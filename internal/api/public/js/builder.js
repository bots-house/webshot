document.addEventListener("DOMContentLoaded", () => {
  
  let buildUrl = (formEl) => {
    const formData = new FormData(formEl);

    const url = new URL(`${window.location.origin}/screenshot`);
    for (let [k, v] of formData.entries()) {
      const inputEl = formEl.querySelector(`input[name=${k}]`)
      if(v != inputEl.defaultValue) {
        url.searchParams.set(k, v);
      }
    }

    return url.toString()
  }

  let renderMessage = url => {
    return `
      <div id="link-message" class="notification is-info">
        <span class="has-text-weight-bold">URL:</span>
        <a id="link" href="${url}">${url}</a>
      </div>

      <div class="card is-loaded">
        <div class="card-image">
          <figure class="image">
            <img src="${url}" alt="Placeholder image">
          </figure>
        </div>
      </div>
    `
  }

  document.querySelector("form").addEventListener("input", (ev) => {
    const form = ev.currentTarget;

    if (form.checkValidity()) {
      const url = buildUrl(form)
      
      const html = renderMessage(url)
      
      form.querySelector('#link-message-root').innerHTML = html
    } 
  });
});
