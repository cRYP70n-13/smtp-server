const { createTransport } = require('nodemailer');

const transporter = createTransport({
    host: "localhost",
    port: 25,
});

const mailOptions = {
    from: 'example@example.com',
    to: 'example.receiver@example.com',
    subject: `testing SMTP server written in GoLang`,
    text: `This is a test for the server that we have written in Go`
};

transporter.sendMail(mailOptions, (error, info) => {
    if (error) {
        console.error(error);
        throw error;
    }

    console.log('Email sent: ' + info.response);
});
