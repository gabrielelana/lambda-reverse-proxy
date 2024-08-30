exports.handler = function (event, context) {
  context.succeed({
    statusCode: 200,
    body: JSON.stringify({
      message: "Hello World!",
      event: event,
      env: process.env
    })
  });
};
