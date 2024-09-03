exports.handler = function (event, context) {
  setTimeout(() => {
    context.succeed({
      statusCode: 200,
      body: JSON.stringify({
        message: "Hello World! (with delay)",
        event: event,
        env: process.env
      })
    });
  }, 300);
};
