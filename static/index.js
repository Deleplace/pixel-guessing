// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Expecting namedParam to have the form
// either "sample=samples/sample18.jpg"
// or     "imgid=y71q"
// so the same code works with either a sample picture, or a user-provided picture.
function startGuessing(namedParam, naturalWidth) {
    console.log("Start guessing", namedParam, "having original width", naturalWidth);
    aiconversation.innerHTML = `
        <tr>
            <th> </th>
            <th class="prompt">
                What does this picture look like? Provide a short answer in less than 8 words.
            </th>
        </tr>
        `;
    let i = 0;
    for(let pixelW=8; pixelW<=naturalWidth; pixelW*=1.3) {
        let w = Math.floor(pixelW);
        setTimeout((namedParam, i) => {
            let row = document.createElement('tr');
            row.innerHTML = `
                <td>
                    <span class="resolution" id="resolution-${i}">${w}</span>
                    <img src="/resized?${namedParam}&pixelwidth=${w}" id="resized-${i}" />
                </td>
                <td class="answer" id="answer-${i}">
                    <div class="loading">Let me guess... </div>
                    <div class="loading">(asking the Gemini Pro Vision model with Vertex AI)</div>
                </td>
            `;
            aiconversation.appendChild(row);
            let img = document.getElementById(`resized-${i}`);
            let resolution = document.getElementById(`resolution-${i}`);
            img.onload = () => {
                // When the pixelated image arrives, show its resolution w x h (in pixels)
                resolution.innerHTML = `${img.naturalWidth}x${img.naturalHeight}`;
            };
            guessSingle(`/guess?${namedParam}&pixelwidth=${w}`, i);
        }, 2000 * i, namedParam, i);
        i++;
    }
}

function guessSingle(endpoint, i) {
    fetch(endpoint)
    .then(response => {
        if (!response.ok) {
            throw new Error("HTTP error " + response.status + ": " + response.text());
        }
        return response.json();
    })
    .then(data => {
        document.getElementById(`answer-${i}`).innerHTML = data.answer;
    })
    .catch(err => {
        document.getElementById(`answer-${i}`).innerHTML = `<span class="error">${err}</span>`;
    });
}

function elementFromHTML(htmlString) {
    var div = document.createElement('div');
    div.innerHTML = htmlString.trim();
    return div.firstChild;
}

// Fisher–Yates
// Thanks to https://stackoverflow.com/a/6274381/871134
function shuffle(a) {
    var j, x, i;
    for (i = a.length - 1; i > 0; i--) {
        j = Math.floor(Math.random() * (i + 1));
        x = a[i];
        a[i] = a[j];
        a[j] = x;
    }
    return a;
}

// Display m sample images, out of n
function selectRandomImages(n, m) {
    let indexes = [];
    for(let i=0; i<n; i++) {
        // Samples indices start at 1
        indexes.push(1 + i);
    }
    shuffle(indexes);
    for(let i=0; i<m; i++) {
        let k = indexes[i];
        let li = document.createElement('li');
        li.innerHTML = `<img src="samples/sample${k}.jpg" alt="sample picture ${k}" />`;
        samples.appendChild(li);
    }
}

up.addEventListener('change', () => {
    let file = up.files[0];
    fetch('/upload', {
        method: 'POST',
        body: file
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("HTTP error " + response.status + ": " + response.text());
        }
        return response.json();
    })
    .then(data => {
        startGuessing("imgid=" + data.imageID, data.width);
    })
    .catch(err => {
        // TODO
        console.error(err);
    });
});

//
// Page init:
//

selectRandomImages(30, 8);

const images = document.querySelectorAll('img');
const sampleImages = document.querySelectorAll('ul.samples > li > img');

sampleImages.forEach(image => {
    image.addEventListener('click', () => {
        images.forEach(otherImage => {
            otherImage.classList.remove('selected');
        });
        image.classList.add('selected');
        let sample = image.getAttribute('src');

        startGuessing("sample=" + sample, image.naturalWidth);
    });
});