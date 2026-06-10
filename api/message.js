let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: {
    bodyParser: false
  }
};

export default async function handler(req, res) {
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  if (req.method === "POST") {
    const contentType = req.headers["content-type"] || "";

    // Lire tout le flux brut
    const chunks = [];
    for await (const chunk of req) chunks.push(chunk);
    const buffer = Buffer.concat(chunks);

    // Si c'est un multipart/form-data
    if (contentType.includes("multipart/form-data")) {
      const boundary = "--" + contentType.split("boundary=")[1];
      const parts = buffer.toString("binary").split(boundary);

      for (const part of parts) {
        if (part.includes("Content-Disposition")) {
          // Message texte
          if (part.includes('name="message"')) {
            const value = part.split("\r\n\r\n")[1]?.split("\r\n")[0];
            if (value) lastMessage = value;
          }

          // Fichier
          if (part.includes("filename=")) {
            const filename = part.match(/filename="(.+?)"/)?.[1];
            const fileContent = part.split("\r\n\r\n")[1];
            const binary = fileContent?.split("\r\n")[0];

            if (filename && binary) {
              lastFile = {
                filename,
                base64: Buffer.from(binary, "binary").toString("base64")
              };
            }
          }
        }
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile ? lastFile.filename : null
      });
    }

    return res.status(400).json({ error: "Format non supporté" });
  }

  return res.status(405).json({ error: "405" });
}
