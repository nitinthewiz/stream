if (localStorage.theme === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark')
} else {
    document.documentElement.classList.remove('dark')
}

var inMemoryPosts;
var currentlyEditingPostID = 0;

fetch('/posts')
    .then((response) => response.json())
        .then((data) => {
            // console.log(data);
            // https://catalins.tech/store-array-in-localstorage/
            localStorage.posts = JSON.stringify(data);
            inMemoryPosts = data;
            for (post of data) {
                // console.log(post.ID);
                // console.log(post.content);
                // console.log(post.contentHTML);
                if (post.contentHTML && post.contentHTML !== '') {
                    document.getElementById(post.ID).innerHTML = post.contentHTML;
                }
                // else {
                //     document.getElementById(post.ID).innerHTML = post.content;
                // }
                if (moment()) {
                    var date_data = new Date(document.getElementById('date-'+post.ID).innerHTML);
                    document.getElementById('date-'+post.ID).innerHTML = moment(date_data).fromNow();
                }
            }
        });



document.querySelector('.submitButton')?.addEventListener('click', handleSubmit);
document.querySelector('.toggleDarkMode')?.addEventListener('click', handleDarkModeToggle);
// document.querySelector('.editButton')?.addEventListener('click', handleEdit);

// https://carlanderson.xyz/adding-event-listeners-to-dynamic-content-with-event-delegation/
document.querySelector('.postsCol').addEventListener('click', event => {
    // console.log("clicked in .postsCol");
        // Check if the clicked element was actually an .editButton
        // if (event.target.matches('.editButton') || event.target.closest('.editButton')) {
        if (event.target.matches('.editButton')) {
        handleEdit(event.target);
        }
});

// document.querySelectorAll('.editButton')?.forEach(editButton => {
//  editButton.addEventListener('click', handleEdit);
// });

async function handleSubmit() {
    var fetchEndpoint = '/posts';
    var fetchMethod = 'POST';
    if (document.querySelector('.textareaClass').value == ''){
        return;
    }

    var data = {
            "author": "me",
            "content": document.querySelector('.textareaClass').value
    }
    if (currentlyEditingPostID != 0){
        fetchEndpoint = fetchEndpoint + '/' + currentlyEditingPostID.toString();
        fetchMethod = 'PUT';
    }
    const response = await fetch(fetchEndpoint, {
                        method: fetchMethod,
                        headers: {
                            'Content-Type': 'application/json',
                            'Token': localStorage.current_secret
                        },
                        body: JSON.stringify(data)
                    });
    if (!response.ok) {
    }
    const body = await response.text();
    document.querySelector('.textareaClass').value = '';
    currentlyEditingPostID = 0;
    location.reload();
}

async function handleEdit(element){
    var foundEditingPostKey = -1;
    // console.log(element);
    // console.log(element.dataset.id);
    // https://developer.mozilla.org/en-US/docs/Learn/HTML/Howto/Use_data_attributes
    currentlyEditingPostID = parseInt(element.dataset.id);
    if (document.querySelector('.textareaClass').value == ''){
        document.querySelector('.textareaClass').value = '';
    }
    // console.log(inMemoryPosts);
    if (inMemoryPosts == null || inMemoryPosts == ""){
        inMemoryPosts = JSON.parse(localStorage.posts);
    }
    for (key in inMemoryPosts) {
        // console.log(post);
        if (inMemoryPosts[key].ID == currentlyEditingPostID){
            foundEditingPostKey = key;
            break;
        }
    }
    if (foundEditingPostKey == -1){
        alert("Could not locate the post you're trying to edit. Maybe refresh the page before trying again.");
    } else {
        post = inMemoryPosts[foundEditingPostKey];
        // console.log(post);
        document.querySelector('.textareaClass').value = post.content;
        // document.querySelector('.textareaClass').scrollIntoView();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }
}

async function handleDarkModeToggle(){
    if (localStorage.theme === 'dark') {
        localStorage.theme = 'light';
    } else {
        localStorage.theme = 'dark';
    }
    location.reload();
}
