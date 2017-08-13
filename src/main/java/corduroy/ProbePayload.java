package corduroy;

import lombok.Getter;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

/**
 * ProbePayload is a payload for discovering a list of node addresses.
 */
public class ProbePayload implements Serializable {

    /**
     * Creates a ProbePayload by initializing the list.
     */
    public ProbePayload() {
        addresses = new ArrayList<String>();
    }

    /**
     * THe list of known node addresses.
     */
    @Getter
    private List<String> addresses;
}
