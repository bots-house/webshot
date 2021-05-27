document.addEventListener("DOMContentLoaded", () => {
  let buildUrl = (formEl) => {
    const formData = new FormData(formEl);

    const url = new URL(`${window.location.origin}/screenshot`);
    for (let [k, v] of formData.entries()) {
      if (k == "key") {
        continue;
      }
      const inputEl = formEl.querySelector(`input[name=${k}]`);
      if (v != inputEl.defaultValue) {
        url.searchParams.set(k, v);
      }
    }


    const keyEl = formEl.querySelector("input[name=key]")
    if (keyEl && keyEl.value) {
      const signUrl = new URL(url);

      const msg = Array.from(signUrl.searchParams.entries())
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map((x) => `${x[0]}=${x[1]}`)
        .join("|");

      url.searchParams.set("sign", sha256.hmac(el.value, msg));
    }

    return url.toString();
  };

  let renderMessage = (url) => {
    return `
      <div id="link-message" class="notification is-info">
        <span class="has-text-weight-bold">URL:</span>
        <a id="link" href="${url}">${url}</a>
      </div>
    `;
  };

  document.querySelector("form").addEventListener("input", (ev) => {
    const form = ev.currentTarget;

    if (form.checkValidity()) {
      const url = buildUrl(form);

      const html = renderMessage(url);

      form.querySelector("#link-message-root").innerHTML = html;
    }
  });
});
