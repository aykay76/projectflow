## Project Selector Test Instructions

### Expected Workflow:
1. **Click the gear icon (⚙️)** in the top-right corner of the header
2. This should open a dropdown menu with "Project" and "Settings" sections
3. **Click the "Select Project" button** in the Project section
4. This should open a project dropdown showing available projects
5. **Click on a project** to select it
6. The header should update to show "Current: [ProjectName]"

### Debugging Steps:

#### Step 1: Check Console Logs
1. Open browser developer tools (F12)
2. Go to Console tab
3. Refresh the page
4. Look for initialization messages:
   - "HeaderMenuManager: Successfully initialized"
   - "Project selector button event listener attached"
   - "Loaded projects: [...]"

#### Step 2: Test Header Menu
1. Click the gear icon (⚙️)
2. Console should show: "Header menu button clicked"
3. Console should show: "HeaderMenuManager: Opening menu"
4. Menu should become visible

#### Step 3: Test Project Selector
1. With menu open, click "Select Project" button
2. Console should show: "Project selector button clicked!"
3. Console should show: "toggleProjectDropdown called"
4. Project dropdown should open

#### Step 4: Check DOM Elements
1. In console, run: `document.getElementById('header-menu-btn')`
2. Should return the button element
3. In console, run: `document.getElementById('project-selector-btn')`
4. Should return the button element
5. In console, run: `document.getElementById('project-dropdown')`
6. Should return the dropdown element

#### Step 5: Manual Test
1. In console, run: 
   ```javascript
   document.getElementById('header-menu').style.display = 'block'
   ```
2. Menu should become visible
3. In console, run:
   ```javascript
   document.getElementById('project-dropdown').style.display = 'block'
   ```
4. Project dropdown should become visible

### If Nothing Happens:
Check these common issues:
1. JavaScript errors in console
2. Missing DOM elements 
3. Event listeners not attached
4. CSS preventing visibility
5. Network issues loading JavaScript files

### Expected Console Output:
```
ProjectManager: setupDOMEventListeners called, document.readyState: complete
HeaderMenuManager: init called, isInitialized: false
HeaderMenuManager: menuBtn element: <button id="header-menu-btn"...>
HeaderMenuManager: menu element: <div id="header-menu"...>
HeaderMenuManager: Binding events
HeaderMenuManager: Successfully initialized
ProjectManager: DOM already ready, attaching listeners
attachProjectDropdownListeners called
Project selector button element: <button id="project-selector-btn"...>
Project selector button event listener attached
Loading available projects...
Loaded projects: [array of projects]
```
