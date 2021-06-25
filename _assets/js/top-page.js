'use strict'

const main = () => {
  readJSON();
}

const readJSON = () => {
  const file = '/data/post-data.json';
  const xhr = new XMLHttpRequest();

  xhr.onreadystatechange = () => {
    if (xhr.readyState == 4 && xhr.status == 200) {
      if (xhr.response) {
        const jsonObj = JSON.parse(xhr.responseText);
        dispatch(jsonObj);
      }
    }
  }
  xhr.open('GET', file, true);
  xhr.send(null);
}

const dispatch = (obj) => {
  const div = document.createElement('div');
  const ul = document.createElement('ul');
  obj.forEach(v => {
    const li = document.createElement('li');
    const link = document.createElement('a');
    link.setAttribute('href', "/".concat(v.Slug, "/"));
    link.innerHTML = v.Title;
    li.appendChild(link);
    ul.appendChild(li);
  });
  div.appendChild(ul);
  document.getElementById('main').appendChild(div);
}

document.addEventListener('DOMContentLoaded', main);

