// Fetch the JSON data
fetch("sample.json")
  .then((response) => response.json())
  .then((data) => {
    const testList = document.getElementById("test-list");
    const tagFilter = document.getElementById("tag-filter");
    const allTags = new Set();

    // Collect all unique tags
    data.forEach((file) => {
      file.tests.forEach((test) => {
        if (test.block.tags) {
          test.block.tags.forEach((tag) => {
            allTags.add(`${tag.key}:${tag.value}`);
          });
        }
      });
    });

    // Create tag filter buttons
    const tagButtons = Array.from(allTags).map((tag) => {
      const [key, value] = tag.split(":");
      return { key, value };
    });

    tagButtons.sort((a, b) => {
      const order = ["team", "module", "info"];
      const indexA = order.indexOf(a.key);
      const indexB = order.indexOf(b.key);

      if (indexA !== indexB) {
        return indexA - indexB;
      } else {
        return a.value.localeCompare(b.value);
      }
    });

    tagButtons.forEach((tag) => {
      const tagButton = document.createElement("button");
      tagButton.classList.add("tag-button");
      tagButton.setAttribute("data-key", tag.key);
      tagButton.textContent = `#${tag.key}:${tag.value}`;
      tagButton.addEventListener("click", () => {
        tagButton.classList.toggle("active");
        filterTests();
      });
      tagFilter.appendChild(tagButton);
    });

    // Filter tests based on selected tags
    function filterTests() {
      const selectedTags = Array.from(tagFilter.querySelectorAll(".tag-button.active")).map((button) =>
        button.textContent.slice(1),
      );

      testList.innerHTML = "";

      data.forEach((file) => {
        const filteredTests = file.tests.filter((test) => {
          if (selectedTags.length === 0) return true;
          if (!test.block.tags) return false;
          const testTags = test.block.tags.map((tag) => `${tag.key}:${tag.value}`);
          return selectedTags.every((tag) => testTags.includes(tag));
        });

        if (filteredTests.length > 0) {
          const fileItem = document.createElement("div");
          fileItem.classList.add("file-item");
          fileItem.innerHTML = `<h2>${file.filename}</h2>`;

          const testListContainer = document.createElement("div");
          testListContainer.classList.add("test-list");

          filteredTests.forEach((test) => {
            const testItem = document.createElement("div");
            testItem.classList.add("test-item");
            testItem.innerHTML = `
              <h3>${test.method}</h3>
              <div class="tags">
                ${
                  test.block.tags
                    ? test.block.tags
                        .sort((a, b) => {
                          const order = ["team", "module", "info"];
                          return order.indexOf(a.key) - order.indexOf(b.key);
                        })
                        .map((tag) => `<span data-key="${tag.key}">#${tag.key}:${tag.value}</span>`)
                        .join("")
                    : ""
                }
              </div>
            `;

            const testDetails = document.createElement("div");
            testDetails.classList.add("test-details");
            testDetails.innerHTML = `
              <p><strong>Given:</strong> ${test.block.given ? test.block.given.join(", ") : "N/A"}</p>
              <p><strong>When:</strong> ${test.block.when ? test.block.when.join(", ") : "N/A"}</p>
              <p><strong>Then:</strong> ${test.block.then ? test.block.then.join(", ") : "N/A"}</p>
            `;

            testItem.appendChild(testDetails);
            testListContainer.appendChild(testItem);
          });

          fileItem.appendChild(testListContainer);
          testList.appendChild(fileItem);
        }
      });
    }

    // Initial rendering of all tests
    filterTests();
  })
  .catch((error) => {
    console.error("Error fetching JSON data:", error);
  });
