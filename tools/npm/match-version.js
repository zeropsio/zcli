const { Toolkit } = require('actions-toolkit')
const { execSync } = require('child_process')


// Run your GitHub Action!
Toolkit.run(async tools => {
  const pkg = tools.getPackageJSON()
  const event = tools.context.payload

  console.log(event);

});
