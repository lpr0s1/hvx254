let lastMessage = "Aucun message pour le moment";
let lastFile = null;

export const config = {
  api: { bodyParser: false }
};

export default async function handler(req, res) {
  // Télécharger le fichier directement
  if (req.method === "GET" && req.url.endsWith("/file")) {
    if (!lastFile) {
      return res.status(404).send("Aucun fichier stocké");
    }

    const buffer = Buffer.from(lastFile.base64, "base64");
    res.setHeader("Content-Type", "application/octet-stream");
    res.setHeader("Content-Disposition", `attachment; filename="${lastFile.filename}"`);
    return res.send(buffer);
  }

  // GET normal
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage,
      file: lastFile
    });
  }

  // POST (envoi fichier + message)
  if (req.method === "POST") {
    try {
      const contentType = req.headers["content-type"] || "";

      if (!contentType.includes("multipart/form-data")) {
        return res.status(400).json({ error: "Format non supporté" });
      }

      const boundaryPart = contentType.split("boundary=")[1];
      if (!boundaryPart) {
        return res.status(400).json({ error: "Boundary manquant" });
      }

      const boundary = "--" + boundaryPart;

      const chunks = [];
      for await (const chunk of req) chunks.push(chunk);
      const buffer = Buffer.concat(chunks);

      const raw = buffer.toString("binary");
      const parts = raw.split(boundary);

      for (const part of parts) {
        if (!part.includes("Content-Disposition")) continue;

        const sections = part.split("\r\n\r\n");
        if (sections.length < 2) continue;

        const header = sections[0];
        let body = sections[1];

        // enlever le suffixe \r\n-- si présent
        body = body.replace(/\r\n--$/, "");

        // message texte
        if (header.includes('name="message"')) {
          lastMessage = body;
        }

        // fichier
        if (header.includes("filename=")) {
          const match = header.match(/filename="(.+?)"/);
          const filename = match ? match[1] : "fichier.bin";

          const fileBuffer = Buffer.from(body, "binary");

          lastFile = {
            filename,
            base64: fileBuffer.toString("base64")
          };
        }
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage,
        receivedFile: lastFile ? lastFile.filename : null
      });
    } catch (err) {
      return res.status(500).json({ error: "Erreur interne", details: err.message });
    }
  }

  return res.status(405).json({ error: "405" });
}
