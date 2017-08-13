package corduroy;

import lombok.Getter;
import lombok.SneakyThrows;

import java.io.ObjectOutputStream;

/**
 * Handler responds to a specific request on behalf of a node.
 */
public class Handler implements Runnable {

    /**
     * Creates a handler by passing in some context about the request.
     * @param requestMessage The request message that was passed in.
     * @param responseStream The response stream that can be written to.
     * @param node The node, for necessary context (eg. to lookup the local table or finger table).
     */
    public Handler(Message requestMessage, ObjectOutputStream responseStream, Node node) {

        this.requestMessage = requestMessage;
        this.responseStream = responseStream;
        this.node = node;
    }

    /**
     * Runs a node by calling handle so the handler can be kicked off in a different thread.
     */
    public void run() { handle(); }

    /**
     * Handles a request and sends a response.
     */
    @SneakyThrows
    public void handle() {

        if (requestMessage.getType() == String.class) {
            handleText();
        }

        else if (requestMessage.getType() == ProbePayload.class) {
            handleProbe();
        }

        complete = true;
    }

    /**
     * Handles a text response (mostly for testing/debugging).
     */
    @SneakyThrows
    public void handleText() {

        String text = requestMessage.getPayload();
        System.out.println(text);

        Message responseMessage = new Message(text.toUpperCase());
        responseStream.writeObject(responseMessage);
        responseStream.flush();
    }

    /**
     * Handles a probe response by broadcasting to the node finger table and returning known node addresses.
     */
    @SneakyThrows
    public void handleProbe() {

        ProbePayload probePayload = requestMessage.getPayload();
        if (!probePayload.getAddresses().contains(node.getAddress())) {

            probePayload.getAddresses().add(node.getAddress());
            for (String requestAddress : node.getFingerTable().values()) {
                if (!probePayload.getAddresses().contains(requestAddress)) {

                    Message responseMessage = Node.send(requestMessage, requestAddress);
                    probePayload = responseMessage.getPayload();
                }
            }
        }

        Message responseMessage = new Message(probePayload);
        responseStream.writeObject(responseMessage);
        responseStream.flush();
    }

    /**
     * The request message.
     */
    private Message requestMessage;

    /**
     * The output stream to write a response to.
     */
    private ObjectOutputStream responseStream;

    /**
     * The node, for context.
     */
    private Node node;

    /**
     * Whether the request handling is complete.
     */
    @Getter
    private boolean complete = false;
}
