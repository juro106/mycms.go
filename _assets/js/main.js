'use strict'

const main = () => {
  const tagName = decodeURI(location.pathname.split('/').reverse()[1]);
  tags.indexOf(tagName) !== -1 && readJSON(tagName);
}

const readJSON = (tag) => {
  const file = '/data/page-data.json';
  const xhr = new XMLHttpRequest();

  xhr.onreadystatechange = () => {
    if (xhr.readyState == 4 && xhr.status == 200) {
      if (xhr.response) {
        const jsonObj = JSON.parse(xhr.responseText);
        dispatch(jsonObj, tag);
      }
    }
  }
  xhr.open('GET', file, true);
  xhr.send(null);
}

const dispatch = (obj, tagName) => {
  const tagList = obj.map(v => {
    console.log(v);
    if (v.Tags && v.Tags.indexOf(tagName) !== null) {
      return {
        slug: v.Slug, 
        title: v.Title
      }
    }
  }).filter(v => v)

  const div = document.createElement('div');
  const h3 = document.createElement('h3');
  h3.innerHTML = 'Links';
  const ul = document.createElement('ul');
  tagList.forEach(v => {
    const li = document.createElement('li');
    const link = document.createElement('a');
    link.setAttribute('href', "/".concat(v.slug, "/"));
    link.innerHTML = v.title;
    li.appendChild(link);
    ul.appendChild(li);
  });
  div.appendChild(ul);
  document.getElementById('main').appendChild(div);
}

document.addEventListener('DOMContentLoaded', main);

